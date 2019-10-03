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

func TrimJsonComma(data []byte, args ...interface{}) ([]byte, error) {
	var (
		e error
		m interface{} = map[string]interface{}{}
	)

	if len(args) > 0 {
		m = args[0]
	}

	if e = json.Unmarshal(data, &m); e != nil {
		tmp := make([]byte, len(data)-1)
		for i := len(data) - 1; i >= 0; i-- {
			if data[i] == ',' {
				copy(tmp, data[:i])
				copy(tmp[i:], data[i+1:])
				if e = json.Unmarshal(tmp, &m); e == nil {
					return tmp, e
				}
			}
		}
	}

	return data, e
}

func TrimJson(data []byte, args ...interface{}) ([]byte, error) {
	var (
		err error
		ret []byte
	)

	if len(args) > 0 {
		ret, err = TrimJsonComma(TrimComment(data), args[0])
	} else {
		ret, err = TrimJsonComma(TrimComment(data))
	}

	return ret, err
}
