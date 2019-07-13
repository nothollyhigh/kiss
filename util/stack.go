package util

import (
	"fmt"
	"runtime"
)

func GetStacks(args ...interface{}) string {
	errstr := ""

	i := 2
	if len(args) > 0 {
		if n, ok := args[0].(int); ok {
			i = n
		}
	}
	for {
		pc, file, line, ok := runtime.Caller(i)

		if !ok || i > 50 {
			break
		}
		errstr += fmt.Sprintf("    stack: %d %v [file: %s] [func: %s] [line: %d]\n", i-1, ok, file, runtime.FuncForPC(pc).Name(), line)

		i++
	}

	return errstr
}
