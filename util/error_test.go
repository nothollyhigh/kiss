package util

import (
	"errors"
	"fmt"
	"testing"
)

func errA() error {
	return ErrorWrap("step 1 failed", errB())
}

func errB() error {
	return ErrorWrap("step 2 failed", errC())
}

func errC() error {
	rawErr := errors.New("step 3 failed")
	return Errorf("%v", rawErr)
}

func TestUtilError(t *testing.T) {
	err1 := errors.New("error1")
	err2 := fmt.Errorf("err2\n%w", err1)
	err3 := fmt.Errorf("err3\n%w", err2)

	fmt.Println(err3)
	fmt.Println(errors.Unwrap(err3))
	fmt.Println(errors.Unwrap(errors.Unwrap(err3)))

	fmt.Println("failed:", errA())
}
