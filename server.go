package dao

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"time"

	"github.com/go-martini/martini"
	"github.com/gorilla/websocket"
)

type wsConn struct {
	ws       *websocket.Conn
	server   *Server
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
	defer handleErrSendCloseChanel(&err)
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

func (conn *wsConn) readRun(hub *WsHub) {
	logger := conn.server.world.logger
	var acc *Account
	defer func() {
		hub.unregister <- conn
		if acc != nil {
			if conn.server.world.IsOnlineAccount(acc) {
				char := acc.UsingChar()
				if char == nil {
					acc.Logout()
				} else {
					char.Logout()
				}
			}
		}
	}()
	conn.ws.SetReadLimit(10240)
	conn.ws.SetReadDeadline(time.Now().Add(60 * time.Second))
	pongFunc := func(string) error {
		conn.ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	}
	conn.ws.SetPongHandler(pongFunc)
ReadLoop:
	for {
		_, msg, err := conn.ws.ReadMessage()
		if err != nil {
			break ReadLoop
		}
		// echo
		// conn.send <- msg
		// TODO
		// connect it to my game
		clientCall := &ClientCall{}
		err = json.Unmarshal(msg, clientCall)
		if err != nil {
			logger.Println(conn.ws.RemoteAddr(), ": can't parse to json:", string(msg))
			continue ReadLoop
		}
		logger.Println(conn.ws.RemoteAddr(), "call:", clientCall)
		switch clientCall.Receiver {
		case "World":
			if acc != nil {
				continue ReadLoop
			}
			v := conn.server.world.WorldClientCall()
			f := reflect.ValueOf(v).MethodByName(clientCall.Method)
			if f.IsNil() {
				continue ReadLoop
			}
			if clientCall.Method == "LoginAccount" {
				clientCall.Params = append(clientCall.Params, conn)
			}
			in, err := clientCall.CastJSON(f)
			if err != nil {
				continue ReadLoop
			}
			if clientCall.Method == "LoginAccount" {
				result := f.Call(in)
				if result[0].IsNil() == false {
					acc = result[0].Interface().(*Account)
				}
			} else {
				f.Call(in)
			}
		case "Account":
			if acc == nil ||
				(acc != nil && acc.UsingChar() != nil) {
				continue ReadLoop
			}
			v := acc.AccountClientCall()
			f := reflect.ValueOf(v).MethodByName(clientCall.Method)
			if f.IsNil() {
				continue ReadLoop
			}
			in, err := clientCall.CastJSON(f)
			if err != nil {
				continue ReadLoop
			}
			f.Call(in)
			if clientCall.Method == "Logout" {
				break ReadLoop
			}
		case "Char":
			if acc == nil {
				continue ReadLoop
			}
			char := acc.UsingChar()
			if char == nil {
				continue ReadLoop
			}
			v := char.CharClientCall()
			f := reflect.ValueOf(v).MethodByName(clientCall.Method)
			if f.IsNil() {
				continue ReadLoop
			}
			in, err := clientCall.CastJSON(f)
			if err != nil {
				continue ReadLoop
			}
			f.Call(in)
			if clientCall.Method == "Logout" {
				break ReadLoop
			}
		default:
			continue ReadLoop
		}
	}
}

type WsHub struct {
	server      *Server
	connections map[*wsConn]struct{}
	register    chan *wsConn
	unregister  chan *wsConn
	quit        chan struct{}
}

func (hub *WsHub) Run() {
	for {
		select {
		case conn := <-hub.register:
			hub.connections[conn] = struct{}{}
		case conn := <-hub.unregister:
			delete(hub.connections, conn)
			close(conn.send)
		case <-hub.quit:
			// TODO
			// should connect with my game
			for conn, _ := range hub.connections {
				close(conn.send)
			}
			hub.quit <- struct{}{}
			return
		}
	}
}

func (hub *WsHub) ShutDown() {
	hub.quit <- struct{}{}
	<-hub.quit
}

type Server struct {
	world *World
	wsHub *WsHub
}

func NewServer() *Server {
	w, err := NewWorld("develop", "127.0.0.1", "dao")
	if err != nil {
		panic(err)
	}
	hub := &WsHub{
		connections: make(map[*wsConn]struct{}),
		register:    make(chan *wsConn),
		unregister:  make(chan *wsConn),
		quit:        make(chan struct{}, 1),
	}
	ds := &Server{
		world: w,
		wsHub: hub,
	}
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
			os.Exit(0)
			break
		}
	}
}

func (s *Server) ShutDown() {
	s.world.ShutDown()
	s.wsHub.ShutDown()
}

func serveWs(w http.ResponseWriter, r *http.Request, ds *Server) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	conn := &wsConn{
		ws:     ws,
		server: ds,
		send:   make(chan []byte, 1024),
	}
	ds.wsHub.register <- conn
	go conn.writeRun()
	conn.readRun(ds.wsHub)
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
	go s.world.Run()
	go s.HandleSignal()
	s.RunHTTP()
	// may be handle pure tcp connection in the future
	// s.RunTCP()
}
