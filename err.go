package dao

import (
	"errors"
	"runtime"
)

var ErrNotSendableChanel = errors.New("err can't send to chanel")

func handleErrSendCloseChanel(err *error) {
	if r := recover(); r != nil {
		err2, ok := r.(runtime.Error)
		if ok && err2.Error() == "runtime error: send on closed channel" {
			*err = ErrNotSendableChanel
		} else {
			panic(r.(error))
		}
	}

}
