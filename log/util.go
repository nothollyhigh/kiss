package log

import (
	"fmt"
	"runtime"
)

const (
	maxStack  = 20
	separator = "---------------------------------------\n"
)

func handlePanic() {
	if err := recover(); err != nil {
		errstr := fmt.Sprintf("\n%sruntime error: %v\ntraceback:\n", separator, err)

		i := 2
		for {
			pc, file, line, ok := runtime.Caller(i)
			if !ok || i > maxStack {
				break
			}
			errstr += fmt.Sprintf("\tstack: %d %v [file: %s] [func: %s] [line: %d]\n", i-1, ok, file, runtime.FuncForPC(pc).Name(), line)
			i++
		}
		errstr += separator

		fmt.Println(errstr)
	}
}
