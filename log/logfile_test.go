package log

import (
	//"encoding/json"
	//"fmt"
	"math/rand"
	"testing"
	//"time"
)

func Benchmark_LogFileBytes128(b *testing.B) {
	w := FileWriter{
		RootDir:     "./logs/",                              //日志根目录
		DirFormat:   "",                                     //日志子目录格式化规则，如果拆分子目录，设置成这种格式"20060102/"
		FileFormat:  "Benchmark_LogFileBytesA.20060102.log", //日志
		MaxFileSize: 0,                                      //单个日志文件最大size，这是为0则不限制单个日志文件大小
		EnableBufio: false,                                  //是否开启bufio，通常设置为false，不开启，若开启，程序意外退出时可能丢日志
	}
	data := make([]byte, 128)
	rand.Read(data)
	for i := 0; i < b.N; i++ {
		w.Write(data)
	}
}

func Benchmark_LogFileBytes256(b *testing.B) {
	w := FileWriter{
		RootDir:     "./logs/",                              //日志根目录
		DirFormat:   "",                                     //日志子目录格式化规则，如果拆分子目录，设置成这种格式"20060102/"
		FileFormat:  "Benchmark_LogFileBytesB.20060102.log", //日志
		MaxFileSize: 0,                                      //单个日志文件最大size，这是为0则不限制单个日志文件大小
		EnableBufio: false,                                  //是否开启bufio，通常设置为false，不开启，若开启，程序意外退出时可能丢日志
	}
	data := make([]byte, 256)
	rand.Read(data)
	for i := 0; i < b.N; i++ {
		w.Write(data)
	}
}

func Benchmark_LogFileBytes512(b *testing.B) {
	w := FileWriter{
		RootDir:     "./logs/",                              //日志根目录
		DirFormat:   "",                                     //日志子目录格式化规则，如果拆分子目录，设置成这种格式"20060102/"
		FileFormat:  "Benchmark_LogFileBytesC.20060102.log", //日志
		MaxFileSize: 0,                                      //单个日志文件最大size，这是为0则不限制单个日志文件大小
		EnableBufio: false,                                  //是否开启bufio，通常设置为false，不开启，若开启，程序意外退出时可能丢日志
	}
	data := make([]byte, 512)
	rand.Read(data)
	for i := 0; i < b.N; i++ {
		w.Write(data)
	}
}

func Benchmark_LogFileBufioBytes128(b *testing.B) {
	w := FileWriter{
		RootDir:     "./logs/",                                   //日志根目录
		DirFormat:   "",                                          //日志子目录格式化规则，如果拆分子目录，设置成这种格式"20060102/"
		FileFormat:  "Benchmark_LogFileBufioBytesA.20060102.log", //日志
		MaxFileSize: 0,                                           //单个日志文件最大size，这是为0则不限制单个日志文件大小
		EnableBufio: true,                                        //是否开启bufio，通常设置为false，不开启，若开启，程序意外退出时可能丢日志
	}
	data := make([]byte, 128)
	rand.Read(data)
	for i := 0; i < b.N; i++ {
		w.Write(data)
	}
}

func Benchmark_LogFileBufioBytes256(b *testing.B) {
	w := FileWriter{
		RootDir:     "./logs/",                                   //日志根目录
		DirFormat:   "",                                          //日志子目录格式化规则，如果拆分子目录，设置成这种格式"20060102/"
		FileFormat:  "Benchmark_LogFileBufioBytesB.20060102.log", //日志
		MaxFileSize: 0,                                           //单个日志文件最大size，这是为0则不限制单个日志文件大小
		EnableBufio: true,                                        //是否开启bufio，通常设置为false，不开启，若开启，程序意外退出时可能丢日志
	}
	data := make([]byte, 256)
	rand.Read(data)
	for i := 0; i < b.N; i++ {
		w.Write(data)
	}
}

func Benchmark_LogFileBufioBytes512(b *testing.B) {
	w := FileWriter{
		RootDir:     "./logs/",                          //日志根目录
		DirFormat:   "",                                 //日志子目录格式化规则，如果拆分子目录，设置成这种格式"20060102/"
		FileFormat:  "Benchmark_LogFileBufioBytesC.log", //日志
		MaxFileSize: 0,                                  //单个日志文件最大size，这是为0则不限制单个日志文件大小
		EnableBufio: true,                               //是否开启bufio，通常设置为false，不开启，若开启，程序意外退出时可能丢日志
	}
	data := make([]byte, 512)
	rand.Read(data)
	for i := 0; i < b.N; i++ {
		w.Write(data)
	}
}
