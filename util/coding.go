package util

import (
	"github.com/axgle/mahonia"
)

func GBKToUTF8(src string) string {
	return mahonia.NewDecoder("gbk").ConvertString(src)
}

func UTF8ToGBK(src string) string {
	return mahonia.NewEncoder("gbk").ConvertString(src)
}
