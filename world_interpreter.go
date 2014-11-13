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
	world *World
	mutex *sync.Mutex
	vm    *otto.Otto
	Quit  chan struct{}
}

func NewWorldInterpreter(w *World) *WorldInterpreter {
	vm := otto.New()
	vm.Set("dao", w)
	v, _ := vm.ToValue(w.Emitter)
	vm.Set("RandUnixNanoTimeSeed", func(call otto.FunctionCall) otto.Value {
		rand.Seed(time.Now().UTC().UnixNano())
		return otto.UndefinedValue()
	})
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
	return &WorldInterpreter{
		world: w,
		vm:    vm,
		mutex: &sync.Mutex{},
		Quit:  make(chan struct{}),
	}
}

func (wi *WorldInterpreter) ResetVM() {
	w := wi.world
	vm := otto.New()
	vm.Set("w", w)
	vm.Set("world", w)
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
	wi.vm = vm
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
		wi.vm.Run(src)
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

func (wi *WorldInterpreter) Eval(expr string) {
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

func (wi *WorldInterpreter) ReadRun() {
	reader := bufio.NewReader(os.Stdin)
	wi.Printf("> ")
	for {
		select {
		case <-wi.Quit:
			wi.Quit <- struct{}{}
			return
		default:
			input, _ := wi.getExpression(reader)
			wi.world.InterpreterEval <- input
		}
	}
}
