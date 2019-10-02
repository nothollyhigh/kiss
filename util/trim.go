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

func FixJsonComma(data []byte) []byte {
	if json.Valid(data) {
		return data
	}

	tmp := make([]byte, len(data)-1)

	for i, v := range data {
		if v == ',' {
			copy(tmp, data[:i])
			copy(tmp[i:], data[i+1:])
			if json.Valid(tmp) {
				return tmp
			}
		}
	}
	return nil
}

func TrimJson(data []byte) []byte {
	return FixJsonComma(TrimComment(data))
}
