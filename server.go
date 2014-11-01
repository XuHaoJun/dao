package dao

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-martini/martini"
	"github.com/gorilla/websocket"
)

type wsConn struct {
	ws       *websocket.Conn
	hub      *WsHub
	server   *Server
	account  *Account
	readQuit chan struct{}
	send     chan []byte
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

func (conn *wsConn) SendMsg(msg interface{}) (err error) {
	err = conn.SendJSON(msg)
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
	world *World
	wsHub *WsHub
}

func NewServer(needReadConfig bool) *Server {
	var w *World
	var err error
	if needReadConfig {
		w, err = NewWorldByConfig(NewDaoConfigsByConfigFiles())
	} else {
		w, err = NewWorldByConfig(NewDefaultDaoConfigs())
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
		world: w,
		wsHub: hub,
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
		ws:      ws,
		hub:     hub,
		server:  hub.server,
		account: nil,
		send:    make(chan []byte, 20480),
	}
}

func (s *Server) RunHTTP() {
	m := martini.Classic()
	m.Get("/testPage", s.testPage)
	go s.wsHub.Run()
	m.Map(s)
	m.Get("/daows", serveWs)
	m.Run()
}

func (s *Server) testPage() string {
	return `<html><body><script src='//ajax.googleapis.com/ajax/libs/jquery/1.10.2/jquery.min.js'></script>
		 <ul id=messages></ul><form><input id=message><input type="submit" id=send value=Send></form>
		 <script>
		 var c=new WebSocket("ws://" + location.hostname + ":"  + location.port + "/daows");
		 c.onopen = function(){
		   c.onmessage = function(response){
		     console.log(response.data);
		     var newMessage = $('<li>').text(response.data);
		     $('#messages').append(newMessage);
		     $('#message').val('');
		   };
		   $('form').submit(function(){
		     c.send($('#message').val());
		     return false;
		   });
		 }
		 </script></body></html>`
}

func (s *Server) Run() {
	go s.RunHTTP()
	go s.HandleSignal()
	s.world.Run()
	// may be handle pure tcp connection in the future
	// s.RunTCP()
}
