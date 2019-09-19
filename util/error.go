package util

import (
	"fmt"
	"runtime"
)

type Error struct {
	text  string
	stack string
	child interface{}
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	var (
		ok  = true
		err = e
		idx = 1
		str = "\n"
	)

	for ok {
		str += fmt.Sprintf("\tstack: %d %v %v\n", idx, err.stack, err.text)
		if err.child == nil {
			break
		}
		idx++
		tmp, ok := err.child.(*Error)
		if !ok {
			str += fmt.Sprintf("\tstack: %d %v %v\n", idx, err.stack, err.child)
			break
		}
		err = tmp
	}

	return str
}

func Errorf(format string, v ...interface{}) *Error {
	var (
		pc, file, line, _ = runtime.Caller(1)
	)

	return &Error{
		text:  fmt.Sprintf(format, v...),
		child: nil,
		stack: fmt.Sprintf("[file: %s] [func: %s] [line: %d]", file, runtime.FuncForPC(pc).Name(), line),
	}
}

func ErrorWrap(text string, v interface{}) *Error {
	var (
		pc, file, line, _ = runtime.Caller(1)
	)

	return &Error{
		text:  text,
		child: v,
		stack: fmt.Sprintf("[file: %s] [func: %s] [line: %d]", file, runtime.FuncForPC(pc).Name(), line),
	}
}
