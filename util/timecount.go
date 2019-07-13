package util

import (
	"fmt"
	"github.com/nothollyhigh/kiss/log"
	"runtime"
	"strings"
	"time"
)

type TimeCOunt struct {
	Tag  string
	B    time.Time
	E    time.Time
	over bool
}

func (tc *TimeCOunt) TimeUsed(tag string) time.Duration {
	if !tc.over {
		tc.over = true
		tc.E = time.Now()
	}
	return tc.E.Sub(tc.B)
}

func (tc *TimeCOunt) String() string {
	if !tc.over {
		tc.over = true
		tc.E = time.Now()
	}
	return fmt.Sprintf("%s timeuse: %v", tc.Tag, tc.E.Sub(tc.B))
}

func (tc *TimeCOunt) prefix(now time.Time) string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = -1
	} else {
		pos := strings.LastIndex(file, "/")
		if pos >= 0 {
			file = file[pos+1:]
		}
	}

	return strings.Join([]string{now.Format(log.DefaultLogger.Layout), fmt.Sprintf(" [Debug] [%s:%d] ", file, line)}, "")
}

func (tc *TimeCOunt) Output() {
	if !tc.over {
		tc.over = true
		tc.E = time.Now()
	}
	log.Println(tc.prefix(tc.E) + tc.String())
}

func (tc *TimeCOunt) Warn(maxTimeUsed time.Duration) {
	if !tc.over {
		tc.over = true
		tc.E = time.Now()
	}

	if tc.E.Sub(tc.B) > maxTimeUsed {
		log.Println(tc.prefix(tc.E) + fmt.Sprintf("%v cost too much time: %v", tc.Tag, tc.E.Sub(tc.B)))
	}
}

func NewTimeCount(tag string) *TimeCOunt {
	return &TimeCOunt{Tag: tag, B: time.Now()}
}
