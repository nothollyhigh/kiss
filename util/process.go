package util

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func GetProcName() string {
	name := os.Args[0]
	if strings.Contains(name, "/") {
		substrs := strings.Split(name, "/")
		name = substrs[len(substrs)-1]
	} else if strings.Contains(name, "\\") {
		substrs := strings.Split(name, "\\")
		name = substrs[len(substrs)-1]
	}

	return name
}

func ProcExist(name string) bool {
	//runtime.GOOS, runtime.GOARCH
	pid := fmt.Sprintf("%d", os.Getpid())
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist")
	} else {
		cmd = exec.Command("ps", "-a")
	}

	buf, err := cmd.Output()
	if err != nil {
		return true
	}

	list := string(buf)
	lines := strings.Split(list, "\n")
	for _, line := range lines {
		if strings.Contains(line, name) && !strings.Contains(line, pid) {
			return true
		}
	}

	return false
}

func CheckSingleProc() {
	if ProcExist(GetProcName()) {
		os.Exit(0)
	}
}
