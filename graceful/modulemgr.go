package graceful

import (
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/util"
	"reflect"
	"sync"
	"time"
)

var (
	defaultModuleMgr = &ModuleMgr{}
)

type M interface {
	Start(args ...interface{})
	After(to time.Duration, f func())
	Stop()
	Push(f func(), args ...interface{}) error
}

type ModuleMgr struct {
	sync.Mutex
	sync.WaitGroup
	modules []M
}

func (mgr *ModuleMgr) Register(m M) {
	mgr.Lock()
	defer mgr.Unlock()

	mgr.modules = append(mgr.modules, m)
}

func (mgr *ModuleMgr) Start(args ...interface{}) {
	mgr.Lock()
	defer mgr.Unlock()

	for _, v := range mgr.modules {
		t := reflect.TypeOf(v)
		log.Debug("Module [%v] Start", t)
		v.Start(args...)
	}
}

func (mgr *ModuleMgr) Stop() {
	log.Debug("ModuleMgr Stop...")
	mgr.Lock()
	defer mgr.Unlock()
	for _, v := range mgr.modules {
		mgr.Add(1)
		module := v
		t := reflect.TypeOf(module)
		log.Debug("Module [%v] Stop...", t)
		util.Go(func() {
			defer mgr.Done()
			defer log.Debug("Module [%v] Stop Done.", t)
			module.Stop()
		})
	}
	mgr.Wait()
	log.Debug("ModuleMgr Stop Done.")
}

func Register(m M) {
	defaultModuleMgr.Register(m)
}

func Start(args ...interface{}) {
	defaultModuleMgr.Start(args...)
}

func Stop() {
	defaultModuleMgr.Stop()
}
