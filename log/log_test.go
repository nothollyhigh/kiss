package log

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func TestLog1(t *testing.T) {
	logPrefix := ""
	timeLayout := DefaultLogTimeLayout

	// 按天切割日志文件，日志根目录下不设子目录，不限制单个日志文件大小
	fileWriter := &FileWriter{
		RootDir:     "./logs1/",     //日志根目录
		DirFormat:   "",             //日志根目录下无子目录
		FileFormat:  "20060102.log", //日志文件命名规则，按天切割文件
		TimeBegin:   len(logPrefix), //解析日志中时间起始位置，用于目录、文件切割，以免日志生成的地方所用时间与logfile写入时间不一致导致的切割偏差
		TimePrefix:  timeLayout,     //解析日志中时间格式
		MaxFileSize: 0,              //单个日志文件最大size，0则不限制size
		EnableBufio: false,          //是否开启bufio
	}

	out := io.MultiWriter(os.Stdout, fileWriter)

	layout := logPrefix + timeLayout
	SetOutput(out)
	SetLevel(LEVEL_INFO)
	SetLogTimeFormat(layout)

	for i := 0; i < 100; i++ {
		Debug(fmt.Sprintf("log %d", i))
		Info(fmt.Sprintf("log %d", i))
		Warn(fmt.Sprintf("log %d", i))
		Error(fmt.Sprintf("log %d", i))
	}
}

// 按天切割日志文件，并限制单个日志文件大小
func TestLog2(t *testing.T) {
	// 按天切割日志文件，日志根目录下子目录按天存储，并限制单个日志文件大小
	// 按天切割日志文件，日志根目录下子目录按天存储，并限制单个日志文件大小
	fileWriter := &FileWriter{
		RootDir:     "./logs2/",     //日志根目录
		DirFormat:   "20060102/",    //日志根目录下按天分割子目录
		FileFormat:  "20060102.log", //日志文件命名规则，按天切割文件
		MaxFileSize: 1024,           //日志文件最大size，按size切割日志文件
		EnableBufio: false,          //是否启用bufio
	}
	out := io.MultiWriter(os.Stdout, fileWriter)

	SetOutput(out)
	SetLevel(LEVEL_WARN)

	for i := 0; i < 100; i++ {
		Debug(fmt.Sprintf("log %d", i))
		Info(fmt.Sprintf("log %d", i))
		Warn(fmt.Sprintf("log %d", i))
		Error(fmt.Sprintf("log %d", i))
	}
}
