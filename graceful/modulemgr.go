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
	// init before module loop
	Init()

	// start module
	Start()

	// stop module, should be graceful
	Stop()

	// Next() set f to execute after timeout and cancel previous f at the same time
	Next(timeout time.Duration, f func())

	// After add f to execute after timeout
	After(timeout time.Duration, f func()) interface{}

	// Cancel timerId which is set by After
	Cancel(timerId interface{})

	// push f to module's queue
	Push(f func(), args ...interface{}) error
}

type ModuleMgr struct {
	sync.Mutex
	sync.WaitGroup
	modules []M
}

// register module to module manager
func (mgr *ModuleMgr) Register(m M) {
	mgr.Lock()
	defer mgr.Unlock()

	mgr.modules = append(mgr.modules, m)
}

// start all modules
func (mgr *ModuleMgr) Start() {
	mgr.Lock()
	defer mgr.Unlock()

	for _, v := range mgr.modules {
		t := reflect.TypeOf(v)
		log.Debug("Module [%v] Start", t)
		v.Init()
		v.Start()
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

func Start() {
	defaultModuleMgr.Start()
}

func Stop() {
	defaultModuleMgr.Stop()
}
