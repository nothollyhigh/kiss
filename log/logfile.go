package log

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// file writer
type FileWriter struct {
	sync.Mutex
	RootDir      string
	DirFormat    string
	FileFormat   string
	TimeBegin    int
	TimePrefix   string
	MaxFileSize  int
	SyncInterval time.Duration
	SaveEach     bool
	EnableBufio  bool

	inited       bool
	currdir      string
	currfile     string
	currFileSize int
	currFileIdx  int

	logfile    *os.File
	filewriter *bufio.Writer
	logticker  *time.Ticker
	inittime   time.Duration

	Formater func(log *Log) string
}

// io.Writer implementation
func (w *FileWriter) Write(p []byte) (n int, err error) {
	w.Lock()
	defer w.Unlock()
	w.checkFileWithData(p)
	if w.EnableBufio {
		n, err = w.filewriter.Write(p)
	} else {
		n, err = w.logfile.Write(p)
	}
	w.currFileSize += n
	if err != nil {
		fmt.Printf("logfile Write failed: %v\n", err)
	}
	if w.SaveEach {
		w.save()
	}
	return n, err
}

// log writer implementation
func (w *FileWriter) WriteLog(log *Log) (n int, err error) {
	w.Lock()
	defer w.Unlock()

	value := log.Value

	if w.Formater != nil {
		value = w.Formater(log)
	} else {
		_, file, line, ok := runtime.Caller(log.Depth + 1)
		if !ok {
			file = "???"
			line = -1
		} else {
			if log.Logger.FullPath {
				for _, v := range filepaths {
					tmp := strings.Replace(file, v, "", 1)
					if tmp != file {
						file = tmp
						break
					}
				}
			} else {
				pos := strings.LastIndex(file, "/")
				if pos >= 0 {
					file = file[pos+1:]
				}
			}
		}

		switch log.Level {
		case LEVEL_PRINT:

		case LEVEL_DEBUG:
			value = strings.Join([]string{log.Time.Format(log.Logger.Layout), fmt.Sprintf(" [Debug] [%s:%d] ", file, line), log.Value, "\n"}, "")
		case LEVEL_INFO:
			value = strings.Join([]string{log.Time.Format(log.Logger.Layout), fmt.Sprintf(" [ Info] [%s:%d] ", file, line), log.Value, "\n"}, "")
		case LEVEL_WARN:
			value = strings.Join([]string{log.Time.Format(log.Logger.Layout), fmt.Sprintf(" [ Warn] [%s:%d] ", file, line), log.Value, "\n"}, "")
		case LEVEL_ERROR:
			value = strings.Join([]string{log.Time.Format(log.Logger.Layout), fmt.Sprintf(" [Error] [%s:%d] ", file, line), log.Value, "\n"}, "")
		case LEVEL_PANIC:
			value = strings.Join([]string{log.Time.Format(log.Logger.Layout), fmt.Sprintf(" [Panic] [%s:%d] ", file, line), log.Value, "\n"}, "")
		case LEVEL_FATAL:
			value = strings.Join([]string{log.Time.Format(log.Logger.Layout), fmt.Sprintf(" [Fatal] [%s:%d] ", file, line), log.Value, "\n"}, "")
		default:
		}
	}

	w.checkFileWithLog(log, len(value))
	if w.EnableBufio {
		n, err = w.filewriter.WriteString(value)
	} else {
		n, err = w.logfile.WriteString(value)
	}
	w.currFileSize += n
	if err != nil {
		fmt.Printf("logfile Write failed: %v\n", err)
	}
	if w.SaveEach {
		w.save()
	}
	return n, err
}

// write string
func (w *FileWriter) WriteString(str string) (n int, err error) {
	w.Lock()
	defer w.Unlock()
	w.checkFileWithString(str)
	if w.EnableBufio {
		n, err = w.filewriter.WriteString(str)
	} else {
		n, err = w.logfile.WriteString(str)
	}
	w.currFileSize += n
	if err != nil {
		fmt.Printf("logfile WriteString failed: %v\n", err)
	}
	if w.SaveEach {
		w.save()
	}
	return n, err
}

// set formater
func (w *FileWriter) SetFormater(f func(log *Log) string) {
	w.Formater = f
}

// flush
func (w *FileWriter) Save() {
	w.Lock()
	defer w.Unlock()
	w.save()
}

// new file
func (w *FileWriter) newFile(path string) error {
	w.logfile = nil
	w.filewriter = nil
	//file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0777)
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err == nil {
		w.logfile = file
		if w.EnableBufio {
			if w.filewriter == nil {
				w.filewriter = bufio.NewWriter(file)
			} else {
				w.filewriter.Reset(file)
			}
		}
	} else {
		fmt.Printf("logfile newFile failed: %s, %s\n", path, err.Error())
	}
	return err
}

// check file
func (w *FileWriter) checkFileWithData(data []byte) bool {
	var (
		err      error = nil
		now      time.Time
		filename = ""
		currfile = ""
	)

	if w.TimePrefix != "" {
		now, err = time.Parse(w.TimePrefix, string(data[w.TimeBegin:w.TimeBegin+len(w.TimePrefix)]))
		if err != nil {
			fmt.Printf("logfile time.Parse(%s) failed: %s\n", string(data[:len(w.TimePrefix)]), err.Error())
			return false
		}
	} else {
		now = time.Now()
	}

	filename = now.Format(w.FileFormat)

	if !w.inited {
		w.Init(now)
	}

	currdir := w.RootDir
	if w.DirFormat != "" {
		currdir += now.Format(w.DirFormat)
	}
	if w.currdir != currdir {
		w.currdir = currdir
		err = w.makeDir(currdir)
	}

	if w.currFileIdx == 0 {
		currfile = currdir + filename
	} else {
		currfile = fmt.Sprintf("%s%s.%04d", currdir, filename, w.currFileIdx)
	}

	if w.currfile != currfile {
		w.currFileIdx = 0
		w.currFileSize = 0
		w.currfile = currfile

		// w.save()
		if w.logfile != nil {
			w.logfile.Close()
		}

		err = w.newFile(w.currfile)
	} else if w.MaxFileSize > 0 && w.currFileSize+len(data) > w.MaxFileSize {
		w.currFileIdx++
		w.currFileSize = 0
		w.currfile = fmt.Sprintf("%s%s.%04d", currdir, filename, w.currFileIdx)

		// w.save()
		if w.logfile != nil {
			w.logfile.Close()
		}

		err = w.newFile(w.currfile)
	}

	return err == nil
}

// check file
func (w *FileWriter) checkFileWithLog(log *Log, size int) bool {
	var (
		err      error = nil
		now            = log.Time
		filename       = ""
		currfile       = ""
	)

	if now.IsZero() {
		now = time.Now()
	}

	filename = now.Format(w.FileFormat)

	if !w.inited {
		w.Init(now)
	}

	currdir := w.RootDir
	if w.DirFormat != "" {
		currdir += now.Format(w.DirFormat)
	}
	if w.currdir != currdir {
		w.currdir = currdir
		err = w.makeDir(currdir)
	}

	if w.currFileIdx == 0 {
		currfile = currdir + filename
	} else {
		currfile = fmt.Sprintf("%s%s.%04d", currdir, filename, w.currFileIdx)
	}

	if w.currfile != currfile {
		w.currFileIdx = 0
		w.currFileSize = 0
		w.currfile = currfile

		// w.save()
		if w.logfile != nil {
			w.logfile.Close()
		}

		err = w.newFile(w.currfile)
	} else if w.MaxFileSize > 0 && w.currFileSize+size > w.MaxFileSize {
		w.currFileIdx++
		w.currFileSize = 0
		w.currfile = fmt.Sprintf("%s%s.%04d", currdir, filename, w.currFileIdx)

		// w.save()
		if w.logfile != nil {
			w.logfile.Close()
		}

		err = w.newFile(w.currfile)
	}

	return err == nil
}

// check file
func (w *FileWriter) checkFileWithString(str string) bool {
	var (
		err      error = nil
		now      time.Time
		filename = ""
		currfile = ""
	)

	if w.TimePrefix != "" {
		now, err = time.Parse(w.TimePrefix, str[w.TimeBegin:w.TimeBegin+len(w.TimePrefix)])
		if err != nil {
			fmt.Printf("logfile time.Parse(%s) failed: %s\n", str[:len(w.TimePrefix)], err.Error())
			return false
		}
	} else {
		now = time.Now()
	}

	filename = now.Format(w.FileFormat)

	if !w.inited {
		w.Init(now)
	}

	currdir := w.RootDir
	if w.DirFormat != "" {
		currdir += now.Format(w.DirFormat) //path.Join(w.RootDir, now.Format(w.DirFormat)) //
	}
	if w.currdir != currdir {
		w.currdir = currdir
		err = w.makeDir(currdir)
	}

	if w.currFileIdx == 0 {
		currfile = currdir + filename
	} else {
		currfile = fmt.Sprintf("%s%s.%04d", currdir, filename, w.currFileIdx)
	}

	if w.currfile != currfile {
		w.currFileIdx = 0
		w.currFileSize = 0
		w.currfile = currfile

		// w.save()
		if w.logfile != nil {
			w.logfile.Close()
		}

		err = w.newFile(w.currfile)
	} else if w.MaxFileSize > 0 && w.currFileSize+len(str) > w.MaxFileSize {
		w.currFileIdx++
		w.currFileSize = 0
		w.currfile = fmt.Sprintf("%s%s.%04d", currdir, filename, w.currFileIdx)

		// w.save()
		if w.logfile != nil {
			w.logfile.Close()
		}

		err = w.newFile(w.currfile)
	}

	return err == nil
}

// make dir
func (w *FileWriter) makeDir(path string) error {
	err := os.MkdirAll(path, 0777)
	if err != nil {
		fmt.Printf("logfile makeDir failed: %s, %s\n", path, err.Error())
	}
	return err
}

// init
func (w *FileWriter) Init(now time.Time) {
	if !w.inited {
		w.inited = true
		currdir := w.RootDir
		if w.DirFormat != "" {
			currdir += now.Format(w.DirFormat)
		}
		if err := w.makeDir(currdir); err != nil {
			fmt.Printf("logfile init mkdir(%s) failed: %v\n", w.RootDir, err)
		}

		if !w.SaveEach { // && w.EnableBufio {
			go func() {
				defer func() {
					recover()
				}()
				if w.SyncInterval <= 0 {
					w.SyncInterval = time.Second * 5
				}
				w.logticker = time.NewTicker(w.SyncInterval)
				for {
					_, ok := <-w.logticker.C
					w.Save()
					if !ok {
						return
					}
				}
			}()
		}
	}
}

// flush
func (w *FileWriter) save() {
	if w.EnableBufio {
		if w.filewriter != nil {
			w.filewriter.Flush()
		}
	} else {
		if w.logfile != nil {
			w.logfile.Sync()
		}
	}
}

// func NewLogFile() *FileWriter {
// 	return &FileWriter{
// 		RootDir: "./logs/",      //日志根目录
// 		DirFormat:  "",             //日志子目录格式化规则，如果拆分子目录，设置成这种格式"20060102/"
// 		FileFormat: "20060102.log", //日志
// 	}
// }
