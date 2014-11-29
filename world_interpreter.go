package dao

import (
	"bufio"
	"fmt"
	"github.com/robertkrimen/otto"
	_ "github.com/robertkrimen/otto/underscore"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

type WorldInterpreter struct {
	world      *World
	timers     map[*OttoTimer]*OttoTimer
	timerReady chan *OttoTimer
	mutex      *sync.Mutex
	vm         *otto.Otto
	QuitRead   chan struct{}
}

type OttoTimer struct {
	timer    *time.Timer
	duration time.Duration
	interval bool
	call     otto.FunctionCall
}

func (wi *WorldInterpreter) NewOttoTimer(call otto.FunctionCall, interval bool) (*OttoTimer, otto.Value) {
	delay, _ := call.Argument(1).ToInteger()
	if 0 >= delay {
		delay = 1
	}

	timer := &OttoTimer{
		duration: time.Duration(delay) * time.Millisecond,
		call:     call,
		interval: interval,
	}
	wi.timers[timer] = timer

	timer.timer = time.AfterFunc(timer.duration, func() {
		wi.world.InterpreterTimer <- timer
	})

	value, err := call.Otto.ToValue(timer)
	if err != nil {
		panic(err)
	}

	return timer, value
}

func NewWorldInterpreter(w *World) *WorldInterpreter {
	wi := &WorldInterpreter{
		world:      w,
		vm:         otto.New(),
		timers:     map[*OttoTimer]*OttoTimer{},
		timerReady: make(chan *OttoTimer, 16),
		mutex:      &sync.Mutex{},
		QuitRead:   make(chan struct{}),
	}
	vm := wi.vm
	vm.Set("dao", w)
	vm.Set("RandUnixNanoTimeSeed", func(call otto.FunctionCall) otto.Value {
		rand.Seed(time.Now().UTC().UnixNano())
		return otto.UndefinedValue()
	})
	v, _ := vm.ToValue(w.Emitter)
	vm.Set("On", func(call otto.FunctionCall) otto.Value {
		event := call.Argument(0)
		listener := call.Argument(1)
		e, err := event.Export()
		if err != nil {
			return otto.NullValue()
		}
		if event.IsString() {
			e = e.(string)
		}
		w.Emitter.On(e, listener)
		return v
	})
	vm.Set("setTimeout", func(call otto.FunctionCall) otto.Value {
		_, value := wi.NewOttoTimer(call, false)
		return value
	})
	vm.Set("setInterval", func(call otto.FunctionCall) otto.Value {
		_, value := wi.NewOttoTimer(call, true)
		return value
	})
	clearTimeout := func(call otto.FunctionCall) otto.Value {
		timer, _ := call.Argument(0).Export()
		if timer, ok := timer.(*OttoTimer); ok {
			timer.timer.Stop()
			delete(wi.timers, timer)
		}
		return otto.UndefinedValue()
	}
	vm.Set("clearTimeout", clearTimeout)
	vm.Set("clearInterval", clearTimeout)
	return wi
}

func (wi *WorldInterpreter) RemoveAndStopAllTimer() {
	for timer, _ := range wi.timers {
		timer.timer.Stop()
	}
	wi.timers = map[*OttoTimer]*OttoTimer{}
}

func (wi *WorldInterpreter) anonymousScript(src []byte) string {
	return fmt.Sprintf("(function(){%s})();", string(src))
}

func (wi *WorldInterpreter) loadScripts(dir string, name string) error {
	path := dir + name
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	config := &ScriptLoadConfigs{}
	err = yaml.Unmarshal(file, config)
	if err != nil {
		return err
	}
	for _, script := range config.Scripts {
		src, err := ioutil.ReadFile(dir + script)
		if err != nil {
			return err
		}
		wi.vm.Run(wi.anonymousScript(src))
	}
	for _, imp := range config.Imports {
		impSpl := strings.SplitAfter(imp, "/")
		impDir := impSpl[:len(impSpl)-1]
		impFile := impSpl[len(impSpl)-1]
		subDir := dir + strings.Join(impDir, "")
		err := wi.loadScripts(subDir, impFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func (wi *WorldInterpreter) LoadScripts() error {
	return wi.loadScripts("./scripts/", "scripts_main.yaml")
}

func (wi *WorldInterpreter) Printf(format string, a ...interface{}) {
	wi.mutex.Lock()
	fmt.Printf(format, a...)
	wi.mutex.Unlock()
}

func (wi *WorldInterpreter) Println(a ...interface{}) {
	wi.mutex.Lock()
	fmt.Println(a...)
	wi.mutex.Unlock()
}

func (wi *WorldInterpreter) VM() *otto.Otto {
	return wi.vm
}

func (wi *WorldInterpreter) REPLEval(expr string) {
	if expr == " " || expr == "" {
		wi.Printf("> ")
		return
	}
	value, err := wi.vm.Run(expr)
	if err != nil {
		wi.Println(err)
	} else {
		wi.Println(value)
	}
	wi.Printf("> ")
}

func (wi *WorldInterpreter) TimerEval(timer *OttoTimer) {
	var arguments []interface{}
	if len(timer.call.ArgumentList) > 2 {
		tmp := timer.call.ArgumentList[2:]
		arguments = make([]interface{}, 2+len(tmp))
		for i, value := range tmp {
			arguments[i+2] = value
		}
	} else {
		arguments = make([]interface{}, 1)
	}
	arguments[0] = timer.call.ArgumentList[0]
	_, err := wi.vm.Call(`Function.call.call`, nil, arguments...)
	if err != nil {
		for _, timer := range wi.timers {
			timer.timer.Stop()
			delete(wi.timers, timer)
		}
	}
	if timer.interval {
		timer.timer.Reset(timer.duration)
	} else {
		delete(wi.timers, timer)
	}
}

func (wi *WorldInterpreter) getLine(reader *bufio.Reader) (string, error) {
	line := make([]byte, 0)
	for {
		linepart, hasMore, err := reader.ReadLine()
		if err != nil {
			return "", err
		}
		line = append(line, linepart...)
		if !hasMore {
			break
		}
	}
	return string(line), nil
}

func (wi *WorldInterpreter) isBalanced(str string) bool {
	parens := 0
	squares := 0
	for _, c := range str {
		switch c {
		case '{':
			parens++
		case '}':
			parens--
		case '[':
			squares++
		case ']':
			squares--
		}
	}
	return parens == 0 && squares == 0
}

func (wi *WorldInterpreter) getExpression(reader *bufio.Reader) (string, error) {
	line, err := wi.getLine(reader)
	if err != nil {
		return "", err
	}
	for !wi.isBalanced(line) {
		wi.Printf(">> ")
		nextline, err := wi.getLine(reader)
		if err != nil {
			return "", err
		}
		line += "\n" + nextline
	}
	return line, nil
}

func (wi *WorldInterpreter) Run() {
	go wi.REPLRun()
}

// TODO
// add some edit or history like a little repl console.
func (wi *WorldInterpreter) REPLRun() {
	reader := bufio.NewReader(os.Stdin)
	wi.Printf("> ")
	for {
		select {
		case <-wi.QuitRead:
			wi.QuitRead <- struct{}{}
			return
		default:
			input, _ := wi.getExpression(reader)
			wi.world.InterpreterREPL <- input
		}
	}
}
