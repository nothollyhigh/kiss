package net

import (
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/util"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	rpprof "runtime/pprof"
	"strings"
	"sync"
	"syscall"
	"time"
)

// http middleware
type HttpHandlerWrapper struct {
	sync.WaitGroup
	handler      http.Handler
	over         bool
	pprofEnabled bool
	pprofRoutes  map[string]func(w http.ResponseWriter, r *http.Request)
}

// enable pprof
func (wrapper *HttpHandlerWrapper) EnablePProf(root string) {
	if !strings.HasSuffix(root, "/") {
		root += "/"
	}
	if wrapper.pprofRoutes == nil {
		wrapper.pprofRoutes = map[string]func(w http.ResponseWriter, r *http.Request){}
	}

	wrapper.pprofEnabled = true

	wrapper.pprofRoutes[root+"cmdline"] = pprof.Cmdline
	wrapper.pprofRoutes[root+"profile"] = pprof.Profile
	wrapper.pprofRoutes[root+"symbol"] = pprof.Symbol
	wrapper.pprofRoutes[root+"trace"] = pprof.Trace
	wrapper.pprofRoutes[root+"index"] = pprof.Index
	for _, v := range rpprof.Profiles() {
		wrapper.pprofRoutes[root+v.Name()] = pprof.Handler(v.Name()).ServeHTTP
	}

	for k, _ := range wrapper.pprofRoutes {
		log.Debug("http server init pprof path: %v", k)
	}
}

// serve http
func (wrapper *HttpHandlerWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wrapper.Add(1)
	defer wrapper.Done()
	defer util.HandlePanic()

	if !wrapper.over {
		if wrapper.pprofEnabled {
			if h, ok := wrapper.pprofRoutes[r.URL.Path]; ok {
				h(w, r)
				return
			}
		}
		if wrapper.handler != nil {
			wrapper.handler.ServeHTTP(w, r)
		} else {
			http.Error(w, http.StatusText(404), 404)
		}
	} else {
		http.Error(w, http.StatusText(404), 404)
	}
}

// http server
type HttpServer struct {
	tag       string
	addr      string
	timeout   time.Duration
	listener  net.Listener
	server    *http.Server
	onTimeout func()
}

// http server
func (svr *HttpServer) Server() *http.Server {
	return svr.server
}

// enable pprof
func (svr *HttpServer) EnablePProf(root string) {
	wraper, _ := svr.server.Handler.(*HttpHandlerWrapper)
	wraper.EnablePProf(root)
}

// serve http
func (svr *HttpServer) Serve() {
	log.Debug("[HttpServer %v] Serve On: %v", svr.tag, svr.addr)
	err := svr.server.Serve(svr.listener)
	log.Debug("[HttpServer %v] Exit: %v", svr.tag, err)
}

// serve https
func (svr *HttpServer) ServeTLS(certFile, keyFile string) {
	log.Debug("[HttpServer %v] ServeTLS On: %v", svr.tag, svr.addr)
	err := svr.server.ServeTLS(svr.listener, certFile, keyFile)
	log.Debug("[HttpServer %v] Exit: %v", svr.tag, err)
}

// graceful shutdown
func (svr *HttpServer) Shutdown() error {
	err := svr.listener.Close()
	log.Debug("[HttpServer %v] shutdown waitting...", svr.tag)
	wrapper := svr.server.Handler.(*HttpHandlerWrapper)
	wrapper.over = true
	wrapper.Done()
	timer := time.AfterFunc(svr.timeout, func() {
		log.Error("[HttpServer %v] shutdown timeout(%v)", svr.tag, svr.timeout)
		if svr.onTimeout != nil {
			svr.onTimeout()
		}
	})
	defer timer.Stop()
	wrapper.Wait()
	log.Debug("[HttpServer %v] shutdown done.", svr.tag)
	return err
}

// setting tcp socket option
func (svr *HttpServer) SetSocketOpt(opt *SocketOpt) {
	if opt != nil {
		readTimeout := time.Second * 60
		readHeaderTimeout := time.Second * 60
		writeTimeout := time.Second * 10
		maxHeaderBytes := 1 << 28
		if opt.ReadTimeout > 0 {
			readTimeout = opt.ReadTimeout
		}
		if opt.ReadHeaderTimeout > 0 {
			readHeaderTimeout = opt.ReadHeaderTimeout
		}
		if opt.WriteTimeout > 0 {
			writeTimeout = opt.WriteTimeout
		}
		maxHeaderBytes = opt.MaxHeaderBytes

		l, ok := svr.listener.(*Listener)
		if ok {

			l.opt = opt
		}

		svr.server.ReadTimeout = readTimeout
		svr.server.ReadHeaderTimeout = readHeaderTimeout
		svr.server.WriteTimeout = writeTimeout
		svr.server.MaxHeaderBytes = maxHeaderBytes
	}
}

// http server factory
func NewHttpServer(tag string, addr string, handler http.Handler, to time.Duration, opt *SocketOpt, onTimeout func()) (*HttpServer, error) {
	listener, err := NewListener(addr, opt)
	if err != nil {
		log.Error("NewHttpServer failed: %v", err)
		return nil, err
	}

	wrapper := &HttpHandlerWrapper{
		handler: handler,
	}
	wrapper.Add(1)

	readTimeout := time.Second * 120
	readHeaderTimeout := time.Second * 60
	writeTimeout := time.Second * 120 //pprof default min timeout 30
	maxHeaderBytes := 1 << 28
	if opt != nil {
		if opt.ReadTimeout > 0 {
			readTimeout = opt.ReadTimeout
		}
		if opt.ReadHeaderTimeout > 0 {
			readHeaderTimeout = opt.ReadHeaderTimeout
		}
		if opt.WriteTimeout > 0 {
			writeTimeout = opt.WriteTimeout
		}
		maxHeaderBytes = opt.MaxHeaderBytes
	}

	svr := &HttpServer{
		tag:      tag,
		addr:     addr,
		timeout:  to,
		listener: listener,
		server: &http.Server{
			Handler:           wrapper,
			ReadTimeout:       readTimeout,
			ReadHeaderTimeout: readHeaderTimeout,
			WriteTimeout:      writeTimeout,
			MaxHeaderBytes:    maxHeaderBytes,
		},
		onTimeout: onTimeout,
	}

	return svr, nil
}

// serve http
func ServeHttp(tag string, addr string, handler http.Handler, timeout time.Duration, opt *SocketOpt) {
	svr, err := NewHttpServer(tag, addr, handler, timeout, opt, func() {
		os.Exit(0)
	})
	if err != nil {
		log.Fatal("[HttpServer %v]: Serve failed: %v", tag, err)
	} else {
		util.Go(svr.Serve)
	}

	util.HandleSignal(func(sig os.Signal) {
		if sig == syscall.SIGTERM || sig == syscall.SIGINT {
			svr.Shutdown()
			os.Exit(0)
		}
	})
}

// serve https
func ServeHttps(tag string, addr string, handler http.Handler, timeout time.Duration, opt *SocketOpt, certFile string, keyFile string) {
	svr, err := NewHttpServer(tag, addr, handler, timeout, opt, func() {
		os.Exit(0)
	})
	if err != nil {
		log.Fatal("[HttpServer %v]: ServeTLS failed: %v", tag, err)
	} else {
		util.Go(func() {
			svr.ServeTLS(certFile, keyFile)
		})
	}

	util.HandleSignal(func(sig os.Signal) {
		if sig == syscall.SIGTERM || sig == syscall.SIGINT {
			svr.Shutdown()
			os.Exit(0)
		}
	})
}
