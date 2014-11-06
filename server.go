package dao

import (
	"encoding/json"
	"github.com/go-martini/martini"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"
)

type wsConn struct {
	ws              *websocket.Conn
	hub             *WsHub
	server          *Server
	account         *Account
	readQuit        chan struct{}
	send            chan []byte
	sendClientCall  chan *ClientCall
	sendClientCalls chan []*ClientCall
}

func (conn *wsConn) write(mt int, msg []byte) error {
	conn.ws.SetWriteDeadline(time.Now().Add(20 * time.Second))
	return conn.ws.WriteMessage(mt, msg)
}

func (conn *wsConn) writeRun() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case msg, ok := <-conn.send:
			if !ok {
				conn.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.write(websocket.TextMessage, msg); err != nil {
				return
			}
		case clientCall, ok := <-conn.sendClientCall:
			if !ok {
				conn.write(websocket.CloseMessage, []byte{})
				return
			}
			msg, err := json.Marshal(clientCall)
			if err != nil {
				continue
			}
			if err := conn.write(websocket.TextMessage, msg); err != nil {
				return
			}
		case clientCalls, ok := <-conn.sendClientCalls:
			if !ok {
				conn.write(websocket.CloseMessage, []byte{})
				return
			}
			msg, err := json.Marshal(clientCalls)
			if err != nil {
				continue
			}
			if err := conn.write(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			if err := conn.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (conn *wsConn) Close() {
	conn.ws.Close()
}

func (conn *wsConn) SendJSON(msg interface{}) (err error) {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	conn.send <- jsonMsg
	return
}

func (conn *wsConn) SendClientCall(msg *ClientCall) {
	conn.sendClientCall <- msg
	return
}

func (conn *wsConn) SendClientCalls(msg []*ClientCall) {
	conn.sendClientCalls <- msg
	return
}

func (conn *wsConn) Send(msg []byte) (err error) {
	defer handleErrSendCloseChanel(&err)
	conn.send <- msg
	return
}

func (conn *wsConn) readRun() {
	defer func() {
		conn.hub.unregister <- conn
	}()
	conn.ws.SetReadLimit(10240)
	conn.ws.SetReadDeadline(time.Now().Add(60 * time.Second))
	pongFunc := func(string) error {
		conn.ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	}
	conn.ws.SetPongHandler(pongFunc)
	for {
		_, msg, err := conn.ws.ReadMessage()
		if err != nil {
			return
		}
		conn.server.world.RequestParseClientCall(msg, conn)
	}
}

type WsHub struct {
	server      *Server
	connections map[*wsConn]struct{}
	register    chan *wsConn
	unregister  chan *wsConn
	Quit        chan struct{}
}

func (hub *WsHub) Run() {
	for {
		select {
		case conn := <-hub.register:
			hub.connections[conn] = struct{}{}
		case conn := <-hub.unregister:
			delete(hub.connections, conn)
			if conn.account != nil {
				conn.account.RequestLogout()
			}
		case <-hub.Quit:
			hub.Quit <- struct{}{}
			return
		}
	}
}

type Server struct {
	world   *World
	wsHub   *WsHub
	configs *DaoConfigs
}

func NewServer(readConfig bool) *Server {
	var w *World
	var err error
	var configs *DaoConfigs
	if readConfig {
		configs = NewDaoConfigsByConfigFiles()
		w, err = NewWorldByConfig(configs)
	} else {
		configs = NewDefaultDaoConfigs()
		w, err = NewWorldByConfig(configs)
	}
	if err != nil {
		panic(err)
	}
	hub := &WsHub{
		connections: make(map[*wsConn]struct{}),
		register:    make(chan *wsConn),
		unregister:  make(chan *wsConn),
		Quit:        make(chan struct{}),
	}
	ds := &Server{
		world:   w,
		wsHub:   hub,
		configs: configs,
	}
	w.server = ds
	hub.server = ds
	return ds
}

func (s *Server) HandleSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for {
		select {
		case <-c:
			s.ShutDown()
			break
		}
	}
}

func (s *Server) ShutDown() {
	s.wsHub.Quit <- struct{}{}
	s.world.Quit <- struct{}{}
	<-s.world.Quit
	<-s.wsHub.Quit
	os.Exit(0)
}

func serveWs(w http.ResponseWriter, r *http.Request, ds *Server) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	upgrader := websocket.Upgrader{
		ReadBufferSize:  10240,
		WriteBufferSize: 10240,
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	conn := NewWsConn(ws, ds.wsHub)
	ds.wsHub.register <- conn
	go conn.writeRun()
	conn.readRun()
}

func NewWsConn(ws *websocket.Conn, hub *WsHub) *wsConn {
	return &wsConn{
		ws:              ws,
		hub:             hub,
		server:          hub.server,
		account:         nil,
		send:            make(chan []byte, 8),
		sendClientCall:  make(chan *ClientCall, 1024),
		sendClientCalls: make(chan []*ClientCall, 1024),
	}
}

func (s *Server) RunHTTP() {
	go s.wsHub.Run()
	port := s.configs.ServerConfigs.HttpPort
	m := martini.Classic()
	m.Map(s)
	m.Get("/daows", serveWs)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), m))
}

func (s *Server) Run() {
	go s.world.Run()
	go s.HandleSignal()
	s.RunHTTP()
}
