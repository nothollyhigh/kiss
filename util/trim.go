package util

import (
	"bufio"
	"bytes"
	"strings"
)

func TrimComment(data []byte) []byte {
	var (
		ret    []byte
		reader = bufio.NewReader(bytes.NewReader(data))
	)

	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		for i, v := range line {
			if v != ' ' && v != '\t' && v != '\r' && v != '\n' {
				if !strings.HasPrefix(string(line[i:]), "//") {
					ret = append(ret, line...)
					ret = append(ret, '\n')
				}
				break
			}
		}
	}
	return ret
}
