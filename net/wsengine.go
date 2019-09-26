package net

import (
	"github.com/gorilla/websocket"
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/util"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

type WSServer struct {
	*WSEngine

	// http server
	*HttpServer

	// websocket upgrader
	upgrader *websocket.Upgrader

	// http请求过滤
	requestHandler func(w http.ResponseWriter, r *http.Request) error

	// ws连接成功
	connectHandler func(cli *WSClient, w http.ResponseWriter, r *http.Request) error

	// 连接断开
	disconnectHandler func(cli *WSClient, w http.ResponseWriter, r *http.Request)

	// 当前连接数
	currLoad int64

	// 过载保护,同时最大连接数
	maxLoad int64

	// handlers map[uint32]func(cli *WSClient, cmd uint32, data []byte)

	// all clients
	clients map[*WSClient]struct{}

	// http handler
	httpHandler func(w http.ResponseWriter, r *http.Request)

	// routers
	wsRoutes map[string]func(http.ResponseWriter, *http.Request)
}

// serve http
func (s *WSServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h, ok := s.wsRoutes[r.URL.Path]; ok {
		h(w, r)
		return
	}

	if s.httpHandler != nil {
		s.httpHandler(w, r)
		return
	}

	http.NotFound(w, r)
}

// setting websocket upgrader
func (s *WSServer) SetUpgrader(upgrader *websocket.Upgrader) {
	s.upgrader = upgrader
}

// handle websocket request
func (s *WSServer) onWebsocketRequest(w http.ResponseWriter, r *http.Request) {
	defer util.HandlePanic()

	if s.shutdown {
		http.NotFound(w, r)
		return
	}

	if s.requestHandler != nil && s.requestHandler(w, r) != nil {
		http.NotFound(w, r)
		return
	}

	online := atomic.AddInt64(&s.currLoad, 1)

	if s.maxLoad > 0 && online > s.maxLoad {
		atomic.AddInt64(&s.currLoad, -1)
		http.NotFound(w, r)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var cli = newClient(conn, s.WSEngine)
	s.Lock()
	s.clients[cli] = struct{}{}
	s.Unlock()

	conn.SetReadLimit(s.ReadLimit)

	defer func() {
		s.Lock()
		delete(s.clients, cli)
		s.Unlock()
		atomic.AddInt64(&s.currLoad, -1)

		cli.Stop()

		if s.disconnectHandler != nil {
			s.disconnectHandler(cli, w, r)
		}
	}()

	if s.connectHandler != nil && s.connectHandler(cli, w, r) != nil {
		return
	}

	conn.SetPingHandler(func(string) error {
		if s.ReadTimeout > 0 {
			conn.SetReadDeadline(time.Now().Add(s.ReadTimeout))
		}
		return nil
	})
	conn.SetPongHandler(func(string) error {
		if s.ReadTimeout > 0 {
			conn.SetReadDeadline(time.Now().Add(s.ReadTimeout))
		}
		return nil
	})

	go cli.writeloop()

	cli.readloop()
}

// current load
func (server *WSServer) CurrLoad() int64 {
	return atomic.LoadInt64(&server.currLoad)
}

// max load
func (server *WSServer) MaxLoad() int64 {
	return server.maxLoad
}

// setting max concurrent
func (server *WSServer) SetMaxConcurrent(maxLoad int64) {
	server.maxLoad = maxLoad
}

// client num
func (s *WSServer) ClientNum() int {
	s.Lock()
	defer s.Unlock()
	return len(s.clients)
}

// setting request handler
func (s *WSServer) HandleRequest(h func(w http.ResponseWriter, r *http.Request) error) {
	s.requestHandler = h
}

// setting websocket connection handler
func (s *WSServer) HandleConnect(h func(cli *WSClient, w http.ResponseWriter, r *http.Request) error) {
	s.connectHandler = h
}

// setting websocket disconnected handler
func (s *WSServer) HandleDisconnect(h func(cli *WSClient, w http.ResponseWriter, r *http.Request)) {
	s.disconnectHandler = h
}

// setting websocket router
func (s *WSServer) HandleWs(path string) {
	s.wsRoutes[path] = s.onWebsocketRequest
}

// setting http router
func (s *WSServer) HandleHttp(h func(w http.ResponseWriter, r *http.Request)) {
	s.httpHandler = h
}

// stop all websocket clients
func (s *WSServer) stopClients() {
	s.Lock()
	defer s.Unlock()

	for client, _ := range s.clients {
		client.Stop()
	}
}

// graceful shutdown
func (s *WSServer) Shutdown(timeout time.Duration, cb func(error)) {
	s.Lock()
	shutdown := s.shutdown
	s.shutdown = true
	s.Unlock()
	if !shutdown {
		log.Debug("WSServer Shutdown ...")

		if timeout <= 0 {
			timeout = DefaultShutdownTimeout
		}

		done := make(chan struct{}, 1)
		util.Go(func() {
			s.Wait()

			s.stopClients()

			s.HttpServer.Shutdown()

			done <- struct{}{}
		})

		after := time.NewTimer(timeout)
		defer after.Stop()
		select {
		case <-after.C:
			log.Debug("WSServer Shutdown timeout")
			if cb != nil {
				cb(ErrWSEngineShutdownTimeout)
			}
		case <-done:
			log.Debug("WSServer Shutdown success")
			cb(nil)
		}
	}
}

// websocket server factory
func NewWebsocketServer(tag string, addr string) (*WSServer, error) {
	var err error
	svr := &WSServer{
		WSEngine: NewWebsocketEngine(),
		maxLoad:  DefaultMaxOnline,
		upgrader: &websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		clients:  map[*WSClient]struct{}{},
		wsRoutes: map[string]func(http.ResponseWriter, *http.Request){},
	}

	svr.HttpServer, err = NewHttpServer(tag, addr, svr, time.Second*5, nil, func() {
		os.Exit(-1)
	})

	return svr, err
}
