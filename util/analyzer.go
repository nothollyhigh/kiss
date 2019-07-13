package util

import (
	"fmt"
	"github.com/nothollyhigh/kiss/log"
	"runtime"
	"strings"
	"time"
)

var (
	analyzerDebug = false
)

func SetAnalyzerDebug(d bool) {
	analyzerDebug = d
}

type Analyzer struct {
	Tag          string
	Parent       *Analyzer `json:"-"`
	Children     []*Analyzer
	Limit        time.Duration `json:"-"`
	SLimit       string        `json:"Limit"`
	TBegin       time.Time
	TEnd         time.Time
	TUsed        string
	StackBegin   string
	StackEnd     string
	Data         interface{}
	Expired      bool
	ChildExpired bool
}

func getStackInfo() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = -1
	} else {
		if log.BuildDir != "" {
			if strings.HasPrefix(file, log.BuildDir) {
				file = file[len(log.BuildDir):]
			}
		}
	}
	return fmt.Sprintf("[%s %d]", file, line)
}

func (a *Analyzer) Begin() {
	if a == nil {
		return
	}
	a.StackBegin = getStackInfo()
}

func (a *Analyzer) Done(v ...interface{}) {
	if a == nil {
		return
	}
	a.TEnd = time.Now()
	a.StackEnd = getStackInfo()
	if len(v) > 0 {
		a.Data = v[0]
	}
	// if a.Expired {
	// 	return
	// }
	tused := a.TEnd.Sub(a.TBegin)
	a.TUsed = fmt.Sprintf("%.6f ms", float64(tused.Nanoseconds()/1000000)+float64(tused.Nanoseconds()%1000000)/float64(1000000))
	a.Expired = tused > a.Limit
	if a.Expired {
		//fmt.Println("+++", a.Tag)
		tmp := a.Parent
		for tmp != nil {
			//fmt.Println("---", tmp.Tag)
			//fmt.Println(a.Tag)
			if tmp.ChildExpired {
				return
			}
			tmp.ChildExpired = true
			tmp = tmp.Parent
		}
	}
}

func (a *Analyzer) Report(v ...interface{}) {
	if a == nil {
		return
	}

	a.TEnd = time.Now()
	a.StackEnd = getStackInfo()
	if len(v) > 0 {
		a.Data = v[0]
	}
	tused := a.TEnd.Sub(a.TBegin)
	a.TUsed = fmt.Sprintf("%.6f ms", float64(tused.Nanoseconds()/1000000)+float64(tused.Nanoseconds()%1000000)/float64(1000000))
	a.Expired = tused > a.Limit
	if a.Expired {
		tmp := a.Parent
		for tmp != nil {
			if tmp.ChildExpired {
				return
			}
			tmp.ChildExpired = true
			tmp = tmp.Parent
		}
	}

	if a.Expired || a.ChildExpired {
		fmt.Println(a.Info())
	}
}

func (a *Analyzer) Fork(tag string, limit time.Duration) *Analyzer {
	if a == nil {
		return nil
	}

	analyzer := &Analyzer{
		Tag:    tag,
		Parent: a,
		Limit:  limit,
		SLimit: fmt.Sprintf("%.6f ms", float64(limit.Nanoseconds())/1000000),
		TBegin: time.Now(),
	}
	a.Children = append(a.Children, analyzer)
	return analyzer
}

// func (a *Analyzer) Info(args ...interface{}) (string, bool) {
// 	indent := ""
// 	if len(args) > 0 {
// 		indent = args[0].(string)
// 	}

// 	used := a.TEnd.Sub(a.TBegin)
// 	expired := used > a.Limit
// 	infoStr := fmt.Sprintf("%v[%v] cost: %vus, exp: %v\n", indent, a.Tag, used.Nanoseconds()/1000, expired)
// 	indent += "--"
// 	for _, v := range a.Children {
// 		str, exp := v.Info(indent)
// 		if exp {
// 			expired = exp
// 		}
// 		infoStr += str
// 	}
// 	return infoStr, expired
// }

// func (a *Analyzer) Expired() bool {
// 	// used := a.TEnd.Sub(a.TBegin)
// 	// expired := used > a.Limit
// 	// for _, v := range a.Children {
// 	// 	if v.Expired() {
// 	// 		return true
// 	// 	}
// 	// }
// 	return a.expired
// }

func (a *Analyzer) Info() string {
	if a == nil {
		return ""
	}

	if analyzerDebug {
		data, _ := json.MarshalIndent(a, "", "    ")
		return string(data)
	}
	str, _ := json.MarshalToString(a)
	return str
}

func NewAnalyzer(tag string, limit time.Duration) *Analyzer {
	return &Analyzer{
		Tag:    tag,
		Limit:  limit,
		SLimit: fmt.Sprintf("%.6f ms", float64(limit.Nanoseconds())/1000000),
		TBegin: time.Now(),
	}
}
