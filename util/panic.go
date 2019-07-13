package util

import (
	"fmt"
	"github.com/nothollyhigh/kiss/log"
	"runtime"
)

const (
	maxStack  = 20
	separator = "---------------------------------------\n"
)

var panicHandler func(string)

func OnPanic(h func(string)) {
	panicHandler = func(str string) {
		defer func() {
			recover()
		}()
		h(str)
	}
}

func HandlePanic() {
	if err := recover(); err != nil {
		errstr := fmt.Sprintf("\n%sruntime error: %v\ntraceback:\n", separator, err)

		i := 2
		for {
			pc, file, line, ok := runtime.Caller(i)
			if !ok || i > maxStack {
				break
			}
			errstr += fmt.Sprintf("    stack: %d %v [file: %s] [func: %s] [line: %d]\n", i-1, ok, file, runtime.FuncForPC(pc).Name(), line)
			i++
		}
		errstr += separator

		if panicHandler != nil {
			panicHandler(errstr)
		} else {
			log.Error(errstr)
		}
	}
}

func Safe(cb func()) {
	defer HandlePanic()
	cb()
}
