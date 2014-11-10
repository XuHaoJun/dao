package dao

import (
	"bufio"
	"fmt"
	"github.com/robertkrimen/otto"
	_ "github.com/robertkrimen/otto/underscore"
	"os"
	"sync"
)

type WorldInterpreter struct {
	world *World
	mutex *sync.Mutex
	vm    *otto.Otto
	Quit  chan struct{}
}

func NewWorldInterpreter(w *World) *WorldInterpreter {
	vm := otto.New()
	vm.Set("w", w)
	vm.Set("world", w)
	return &WorldInterpreter{
		world: w,
		vm:    vm,
		mutex: &sync.Mutex{},
		Quit:  make(chan struct{}),
	}
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
