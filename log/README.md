#### 一、日志文件切割

```golang
package main

import (
	"fmt"
	"github.com/nothollyhigh/kiss/log"
	"io"
	"os"
)

func main() {
	fileWriter := &log.FileWriter{
		RootDir:     "./logs/",            //日志根目录，每次启动新建目录则可以 "./logs/" + time.Now().Format("20060102150405") + "/"
		DirFormat:   "200601021504/",      //按时间格式分割日志文件子目录，""则不拆分子目录；此处测试按分钟
		FileFormat:  "20060102150405.log", //按时间格式切割日志文件，此处测试按秒
		MaxFileSize: 1024 * 256,           //按最大size切割日志文件
		EnableBufio: false,                //是否启用bufio，重要日志建议不开启
	}

	out := io.MultiWriter(os.Stdout, fileWriter)
	log.SetOutput(out)

	log.SetLevel(log.LEVEL_WARN)

	i := 0
	for {
		i++
		log.Debug(fmt.Sprintf("log %d", i))
		log.Info(fmt.Sprintf("log %d", i))
		log.Warn(fmt.Sprintf("log %d", i))
		log.Error(fmt.Sprintf("log %d", i))
	}
}
```

#### 二、在gin中使用kiss log
```golang
package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nothollyhigh/kiss/log"
	"os"
	"time"
)

func main() {
	router := gin.New()

	router.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		comment := c.Errors.ByType(gin.ErrorTypePrivate).String()
		if raw != "" {
			path = path + "?" + raw
		}
		logLevel := log.LEVEL_INFO
		if latency > time.Second {
			logLevel = log.LEVEL_WARN
		}

		fmt.Fprintf(os.Stdout, log.LogWithFormater(logLevel, log.DefaultLogDepth-1, log.DefaultLogTimeLayout, "| %3d | %13v | %15s | %-7s %s\n%s",
			statusCode,
			latency,
			clientIP,
			method,
			path,
			comment,
		))
	})

	//gin.SetMode(gin.ReleaseMode)

	n := 0
	router.GET("/hello", func(c *gin.Context) {
		n++
		if n%2 == 1 {
			time.Sleep(time.Second)
		}
		c.String(200, "hello")
	})

	router.Run(":8080")
}
```


#### 三、自定义日志处理

```golang
package main

import (
	"encoding/json"
	"fmt"
	"github.com/nothollyhigh/kiss/log"
	"runtime"
	"strings"
)

var (
	// 按天切割日志文件，日志根目录下不设子目录，不限制单个日志文件大小
	fileWriter = &log.FileWriter{
		RootDir:     "./logs/",        //日志根目录
		DirFormat:   "20060102-1504/", //日志根目录下无子目录
		FileFormat:  "20060102.log",   //日志文件命名规则，按天切割文件
		MaxFileSize: 1024 * 1024,      //单个日志文件最大size，0则不限制size
		EnableBufio: false,            //是否开启bufio
	}
)

// 实现log.ILogWriter接口
type LogWriter struct{}

func (w *LogWriter) WriteLog(l *log.Log) (n int, err error) {
	_, file, line, ok := runtime.Caller(l.Depth + 1)
	if !ok {
		file = "???"
		line = -1
	} else {
		pos := strings.LastIndex(file, "/")
		if pos >= 0 {
			file = file[pos+1:]
		}
	}

	l.File = file
	l.Line = line

	data, _ := json.Marshal(l)
	data = append(data, '\n')
	n, err = fileWriter.Write(data)
	fmt.Printf(string(data))

	return n, err
}

func main() {
	// 设置为nil则默认的日志不处理，如果需要可以同时使用
	log.SetOutput(nil)

	// 设置自定义日志处理接口
	log.SetStructOutput(&LogWriter{})

	log.SetLevel(log.LEVEL_WARN)

	for i := 0; i < 10000; i++ {
		log.Debug("log %d", i)
		log.Info("log %d", i)
		log.Warn("log %d", i)
		log.Error("log %d", i)
	}
}
```