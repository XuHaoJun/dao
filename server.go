package dao

import (
	"encoding/json"
	"github.com/go-martini/martini"
	"github.com/gorilla/websocket"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/gzip"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/xuhaojun/oauth2"
	goauth2 "golang.org/x/oauth2"
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
	conn.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return conn.ws.WriteMessage(mt, msg)
}

func (conn *wsConn) writeRun() {
	ticker := time.NewTicker(50 * time.Second)
	defer func() {
		ticker.Stop()
		conn.ws.Close()
	}()
	for {
		select {
		case <-ticker.C:
			if err := conn.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
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
	conn.ws.SetReadLimit(20480)
	conn.ws.SetReadDeadline(time.Now().Add(60 * time.Second))
	pongFunc := func(string) error {
		conn.ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	}
	conn.ws.SetPongHandler(pongFunc)
	for {
		_, msg, err := conn.ws.ReadMessage()
		if err != nil {
			break
		}
		clientCall := &ClientCall{}
		err = json.Unmarshal(msg, clientCall)
		if err != nil {
			continue
		}
		conn.server.world.RequestParseClientCall(clientCall, conn)
	}
}

type WsHub struct {
	server      *Server
	wsUpgrader  *websocket.Upgrader
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
			} else {
				conn.ws.Close()
			}
		case <-hub.Quit:
			hub.Quit <- struct{}{}
			return
		}
	}
}

type Server struct {
	world   *World
	db      *DaoDB
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
	db, err := NewDaoDB(configs.MongoDBConfigs.URL,
		configs.MongoDBConfigs.DBName)
	if err != nil {
		panic(err)
	}
	hub := &WsHub{
		connections: make(map[*wsConn]struct{}),
		register:    make(chan *wsConn),
		unregister:  make(chan *wsConn),
		wsUpgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		Quit: make(chan struct{}),
	}
	ds := &Server{
		world:   w,
		wsHub:   hub,
		db:      db,
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
	ws, err := ds.wsHub.wsUpgrader.Upgrade(w, r, nil)
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

func (s *Server) RunWebSocket() {
	go s.wsHub.Run()
	r := martini.NewRouter()
	r.Get("/daows", serveWs)
	m := martini.New()
	m.Use(martini.Logger())
	m.Use(martini.Recovery())
	m.MapTo(r, (*martini.Routes)(nil))
	m.Map(s)
	m.Action(r.Handle)
	websocketPort := s.configs.ServerConfigs.WebSocketPort
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(websocketPort), m))
}

func (s *Server) RunWeb() {
	configs := s.configs
	sessionKey := s.configs.ServerConfigs.SessionKey
	store := sessions.NewCookieStore([]byte(sessionKey))
	m := martini.Classic()
	m.Map(s.configs)
	m.Map(s.world)
	m.Use(gzip.All())
	m.Use(sessions.Sessions("_auth", store))
	m.Use(s.DBHandler())
	m.Use(render.Renderer())
	if configs.ServerConfigs.EnableOauth2 {
		fbAuth := configs.Oauth2Configs.Facebook
		oauth2.PathLogin = "/oauth2login"
		oauth2.PathLogout = "/oauth2logout"
		facebookAuth := oauth2.Facebook(
			goauth2.Client(fbAuth.ClientId, fbAuth.ClientSecret),
			goauth2.RedirectURL(fbAuth.RedirectURL),
			goauth2.Scope(fbAuth.Scope),
		)
		m.Use(facebookAuth)
		// create account
		m.Get("/account/newByFacebook", oauth2.LoginRequired, handleAccountRegisterByFacebook)
		// login account
		m.Get("/account/loginByFacebook/:ltype", oauth2.LoginRequired, handleAccountLoginByFacebook)
		m.Get("/account/loginGameByFacebook", oauth2.LoginRequired, handleAccountLoginGameByFacebook)
		m.Get("/account/loginWebByFacebook", oauth2.LoginRequired, handleAccountLoginWebByFacebook)
	}
	// create account
	m.Post("/account", binding.Bind(AccountRegisterFrom{}), handleAccountRegister)
	// login account
	m.Post("/account/login/:ltype", binding.Bind(AccountLoginForm{}), handleAccountLogin)
	m.Post("/account/loginWeb", binding.Bind(AccountLoginForm{}), handleAccountLoginWeb)
	m.Post("/account/loginGame", binding.Bind(AccountLoginForm{}), handleAccountLoginGame)
	m.Get("/account/loginGamebySession", handleAccountLoginGameBySession)
	// logout account
	m.Get("/account/logout", hanldeAccountLogout)
	// get account
	m.Get("/account", handleAccountInfo)
	m.Get("/account/isLogined", haldeAccountIsLogined)
	// server run
	httpPort := s.configs.ServerConfigs.HttpPort
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(httpPort), m))
}

func (s *Server) Run() {
	go s.RunWeb()
	go s.RunWebSocket()
	go s.HandleSignal()
	s.world.Run()
}
