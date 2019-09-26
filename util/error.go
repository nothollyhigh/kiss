package util

import (
	"errors"
	"fmt"
	"runtime"
)

var (
	MAX_ERROR_STACK   = 128
	ERROR_LINE_PREFIX = "\n\t"
)

type Error struct {
	Deep  int
	Text  string
	Stack string
	Child *Error
}

func (e *Error) ToError() error {
	return errors.New(e.Text)
}

func (e *Error) Root() *Error {
	if e == nil {
		return nil
	}

	for i := 0; i < MAX_ERROR_STACK; i++ {
		if e.Child == nil {
			return e
		}
		e = e.Child
	}

	panic(fmt.Errorf("error stack overflow, MAX_ERROR_STACK: %v", MAX_ERROR_STACK))
	return nil
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	var (
		err = e
		str = ""
	)

	for err != nil {
		str += fmt.Sprintf("%v%v %v", ERROR_LINE_PREFIX, err.Stack, err.Text)
		err = err.Child
	}

	return str
}

func Errorf(format string, v ...interface{}) *Error {
	var (
		pc, file, line, _ = runtime.Caller(1)
	)

	return &Error{
		Deep:  1,
		Text:  fmt.Sprintf(format, v...),
		Child: nil,
		Stack: fmt.Sprintf("[file: %s] [func: %s] [line: %d]", file, runtime.FuncForPC(pc).Name(), line),
	}
}

func ErrorWrap(text string, v interface{}) *Error {
	var (
		pc, file, line, _ = runtime.Caller(1)

		ret = &Error{
			Text:  text,
			Stack: fmt.Sprintf("[file: %s] [func: %s] [line: %d]", file, runtime.FuncForPC(pc).Name(), line),
		}
	)

	e, ok := v.(*Error)
	if ok {
		ret.Child = e
	} else {
		ret.Child = Errorf("%v", v)
	}

	ret.Deep = e.Deep + 1

	if ret.Deep > MAX_ERROR_STACK {
		panic(fmt.Errorf("error stack overflow, MAX_ERROR_STACK: %v", MAX_ERROR_STACK))
	}

	return ret
}

func SetMaxErrorStack(max int) {
	MAX_ERROR_STACK = max
}

func SetErrorLinePrefix(prefix string) {
	ERROR_LINE_PREFIX = prefix
}
