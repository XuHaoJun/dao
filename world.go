package dao

import (
	"errors"
	"github.com/xuhaojun/emission-otto"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"os"
	"reflect"
	"sync"
	"time"
)

type AccountLoginBySession struct {
	Username     string
	SessionToken string
}

type World struct {
	*emission.Emitter
	name     string
	server   *Server
	accounts map[string]*Account
	scenes   map[string]*Scene
	timers   map[*WorldTimer]*WorldTimer
	db       *DaoDB
	configs  *DaoConfigs
	logger   *log.Logger
	//
	accountLoginBySessionMap map[string]string
	addAccountLoginBySession chan *AccountLoginBySession
	// SaveAccount     chan *Account
	// RegisterAccount chan *WorldRegisterAccount
	// LoginAccount    chan *WorldLoginAccount
	LogoutAccount chan *Account
	//
	SceneObjecterChangeScene chan *ChangeScene
	//
	BioReborn chan Bioer
	//
	// AccountLoginChar  chan *AccountLoginChar
	// AccountCreateChar chan *AccountCreateChar
	//
	ParseClientCall chan *WorldParseClientCall
	//
	interpreter      *WorldInterpreter
	InterpreterREPL  chan string
	InterpreterTimer chan *OttoTimer
	worldTimer       chan *WorldTimer
	//
	delta    float32
	timeStep time.Duration
	//
	job  chan func()
	Quit chan struct{}
	//
	util  *Util
	cache *Cache
}

type ChangeScene struct {
	SceneObjecter SceneObjecter
	Scene         *Scene
}

type WorldClientCall interface {
	RegisterAccount(username string, password string, email string, sock *wsConn)
	LoginAccount(username string, password string, sock *wsConn)
	LoginAccountBySessionToken(username string, token string, sock *wsConn)
}

func NewWorldByConfig(dc *DaoConfigs) (w *World, err error) {
	w, err = NewWorld(dc.WorldConfigs.Name,
		dc.MongoDBConfigs.URL, dc.MongoDBConfigs.DBName, dc)
	return
}

type WorldTimer struct {
	timer    *time.Timer
	duration time.Duration
	interval bool
	call     reflect.Value
}

func NewWorld(name string, mgourl string, dbname string, configs *DaoConfigs) (*World, error) {
	db, err := NewDaoDB(mgourl, dbname)
	if err != nil {
		return nil, err
	}
	go db.UpdateAccountIndex()
	//
	err = db.ImportDefaultJsonDB()
	if err != nil {
		return nil, err
	}
	w := &World{
		name:                     name,
		accounts:                 make(map[string]*Account),
		scenes:                   make(map[string]*Scene),
		db:                       db,
		configs:                  NewDefaultDaoConfigs(),
		logger:                   log.New(os.Stdout, "[dao-"+name+"] ", 0),
		LogoutAccount:            make(chan *Account, 8),
		SceneObjecterChangeScene: make(chan *ChangeScene, 8),
		ParseClientCall:          make(chan *WorldParseClientCall, 10240),
		InterpreterREPL:          make(chan string, 8),
		InterpreterTimer:         make(chan *OttoTimer, 8),
		worldTimer:               make(chan *WorldTimer, 128),
		timers:                   make(map[*WorldTimer]*WorldTimer),
		BioReborn:                make(chan Bioer, 10240),
		accountLoginBySessionMap: make(map[string]string, 32),
		addAccountLoginBySession: make(chan *AccountLoginBySession, 8),
		//
		delta:    1.0 / 60.0,
		timeStep: (1.0 * time.Second / 60.0),
		//
		Quit: make(chan struct{}),
		//
		util:  &Util{},
		cache: NewCache(),
	}
	if configs != nil {
		w.configs = configs
	}
	// scenes
	daoCity := NewWallScene(w, "daoCity", 2000, 2000)
	w.scenes["daoCity"] = daoCity
	//
	daoField01 := NewWallScene(w, "daoField01", 6000, 6000)
	daoField01.defaultGroundTextureName = "dirt"
	w.scenes["daoField01"] = daoField01
	// interpreter
	w.interpreter = NewWorldInterpreter(w)
	w.Emitter = emission.NewEmitterOtto(w.interpreter.vm)
	err = w.interpreter.LoadScripts()
	if err != nil {
		return nil, err
	}
	// after create scenes
	w.configs.SceneConfigs.SetScenes(w.scenes)
	w.Emit("worldLoadScenes", w, w.scenes)
	//
	return w, nil
}

func (w *World) NewMobByBaseId(id int64) Mober {
	return NewMobByBaseId(w, int(id))
}

func (w *World) NewNpcByBaseId(id int64) Npcer {
	return NewNpcByBaseId(w, int(id))
}

func (w *World) Name() string {
	return w.name
}

func (w *World) ReloadJsonDB() (err error) {
	w.logger.Println("Reloading JsonDB")
	err = w.db.ImportDefaultJsonDB()
	if err != nil {
		w.logger.Println("Error ReloadJsonDB")
		return
	}
	w.cache = NewCache()
	for _, acc := range w.accounts {
		char := acc.usingChar
		if char == nil {
			continue
		}
		char.UpdateItemsUseSelfItemFunc()
	}
	w.logger.Println("Reloaded JsonDB!")
	return
}

func (w *World) ReloadDaoConfigs() (err error) {
	w.logger.Println("Reloading DaoConfigs")
	w.configs.ReloadConfigFiles()
	w.configs.SceneConfigs.SetScenes(w.scenes)
	w.logger.Println("Reloaded DaoConfigs!")
	return
}

func (w *World) ReloadScripts() {
	w.logger.Println("Reloading Scripts")
	// w.interpreter.ResetVM()
	w.Emitter.ResetOttoEvents()
	w.interpreter.RemoveAndStopAllTimer()
	for _, scene := range w.scenes {
		scene.RemoveAllMober()
		scene.RemoveAllNpcer()
	}
	w.interpreter.LoadScripts()
	w.Emit("worldLoadScenes", w, w.scenes)
	w.logger.Println("Reloaded Scripts!")
}

func (w *World) ReloadAll() {
	w.ReloadJsonDB()
	w.ReloadDaoConfigs()
	w.ReloadScripts()
}

func (w *World) WorldClientCall() WorldClientCall {
	return w
}

func (w *World) Run() {
	defer w.db.session.Close()
	var wg sync.WaitGroup
	go w.interpreter.Run()
	physicC := time.Tick(w.timeStep)
	for {
		select {
		case <-physicC:
			for _, scene := range w.scenes {
				wg.Add(1)
				go func(s *Scene, dt float32) {
					s.Update(dt)
					wg.Done()
				}(scene, w.delta)
			}
			wg.Wait()
		case params := <-w.ParseClientCall:
			w.DoParseClientCall(params.ClientCall, params.Conn)
		case expr := <-w.InterpreterREPL:
			w.interpreter.REPLEval(expr)
		case timer := <-w.InterpreterTimer:
			w.interpreter.TimerEval(timer)
		case timer := <-w.worldTimer:
			w.TimerEval(timer)
		case acc := <-w.LogoutAccount:
			w.DoLogoutAccount(acc)
		case params := <-w.addAccountLoginBySession:
			username := params.Username
			sessionToken := params.SessionToken
			w.accountLoginBySessionMap[username] = sessionToken
		case params := <-w.SceneObjecterChangeScene:
			sb := params.SceneObjecter
			scene := params.Scene
			if sb.Scene() == params.Scene {
				continue
			}
			if sb.Scene() == nil {
				scene.Add(sb)
			} else {
				sb.Scene().Remove(sb)
				scene.Add(sb)
			}
		case b := <-w.BioReborn:
			b.Reborn()
		case <-w.Quit:
			for _, acc := range w.accounts {
				acc.Logout()
			}
			w.Quit <- struct{}{}
			return
		}
	}
}

// TODO
// should check username and password is right format!
func (w *World) registerAccount(username string, password string, email string, sock *wsConn) {
	db := w.db
	queryAcc := bson.M{"username": username}
	err := db.accounts.Find(queryAcc).Select(bson.M{"_id": 1}).One(&struct{}{})
	if err != nil && err != mgo.ErrNotFound {
		panic(err)
	} else if err != mgo.ErrNotFound {
		clientErr := []interface{}{"duplicated account!"}
		clientCall := &ClientCall{
			Receiver: "world",
			Method:   "handleErrorLoginAccount",
			Params:   clientErr,
		}
		sock.SendClientCall(clientCall)
	} else if err == mgo.ErrNotFound {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			panic(err)
		}
		acc := NewAccount(username, string(hashedPassword))
		acc.email = email
		acc.world = w
		acc.maxChars = w.configs.AccountConfigs.MaxChars
		acc.Save()
		go w.db.UpdateAccountIndex()
		clientParams := []interface{}{"success register a new account!"}
		clientCall := &ClientCall{
			Receiver: "world",
			Method:   "handleSuccessRegisterAccount",
			Params:   clientParams,
		}
		sock.SendClientCall(clientCall)
	}
}

func (w *World) RegisterAccount(username string, password string, email string, sock *wsConn) {
	go w.registerAccount(username, password, email, sock)
}

func (w *World) ShutDownServer() {
	go w.server.ShutDown()
}

type WorldParseClientCall struct {
	ClientCall *ClientCall
	Conn       *wsConn
}

func (w *World) DoParseClientCall(clientCall *ClientCall, conn *wsConn) {
	acc := conn.account
	switch clientCall.Receiver {
	case "World":
		if acc != nil {
			return
		}
		v := w.WorldClientCall()
		f := reflect.ValueOf(v).MethodByName(clientCall.Method)
		if f.IsValid() == false {
			return
		}
		if clientCall.Method == "LoginAccount" ||
			clientCall.Method == "LoginAccountBySessionToken" ||
			clientCall.Method == "RegisterAccount" {
			clientCall.Params = append(clientCall.Params, conn)
		}
		in, err := clientCall.CastJSON(f)
		if err != nil {
			return
		}
		f.Call(in)
	case "Account":
		if acc == nil {
			return
		}
		v := acc.AccountClientCall()
		f := reflect.ValueOf(v).MethodByName(clientCall.Method)
		if f.IsValid() == false {
			return
		}
		in, err := clientCall.CastJSON(f)
		if err != nil {
			return
		}
		f.Call(in)
	case "Char":
		if acc == nil {
			return
		}
		char := acc.UsingChar()
		if char == nil {
			return
		}
		v := char.CharClientCall()
		f := reflect.ValueOf(v).MethodByName(clientCall.Method)
		if f.IsValid() == false {
			return
		}
		in, err := clientCall.CastJSON(f)
		if err != nil {
			return
		}
		f.Call(in)
	default:
		return
	}
}

func (w *World) RequestParseClientCall(c *ClientCall, conn *wsConn) {
	w.ParseClientCall <- &WorldParseClientCall{c, conn}
}

func (w *World) RemoveAccount(acc *Account) bool {
	_, ok := w.accounts[acc.username]
	if ok {
		delete(w.accounts, acc.username)
	}
	return ok
}

func (w *World) DoLogoutAccount(acc *Account) {
	ok := w.RemoveAccount(acc)
	if ok {
		acc.Logout()
	}
}

func (w *World) KickAccountByUsername(username string) {
	acc, ok := w.accounts[username]
	if ok {
		acc.Logout()
		delete(w.accounts, acc.username)
	}
}

func (w *World) KickAllAccount() {
	for _, acc := range w.accounts {
		acc.Logout()
		delete(w.accounts, acc.username)
	}
}

func (w *World) OnlineAccountUsernames() []string {
	names := make([]string, len(w.accounts))
	i := 0
	for name, _ := range w.accounts {
		names[i] = name
	}
	return names
}

func (w *World) TalkWorld(name string, content string) {
	clientCall := &ClientCall{
		Receiver: "char",
		Method:   "handleChatMessage",
		Params: []interface{}{
			&ChatMessageClient{
				time.Now(),
				"World",
				name,
				content,
			},
		},
	}
	for _, acc := range w.accounts {
		if acc.UsingChar() == nil {
			continue
		}
		char := acc.UsingChar()
		char.sock.SendClientCall(clientCall)
	}
}

func (w *World) LoginAccountBySessionToken(username string, token string, sock *wsConn) {
	_, isOnlineAccount := w.accounts[username]
	realToken, found := w.accountLoginBySessionMap[username]
	if isOnlineAccount || !found || realToken != token {
		clientErr := []interface{}{"wrong username or password"}
		clientCall := &ClientCall{
			Receiver: "world",
			Method:   "handleErrorLoginAccount",
			Params:   clientErr,
		}
		sock.SendClientCall(clientCall)
		return
	}
	delete(w.accountLoginBySessionMap, username)
	foundAcc := &AccountDumpDB{}
	queryAcc := bson.M{"username": username}
	err := w.db.accounts.Find(queryAcc).One(foundAcc)
	if err != nil && err != mgo.ErrNotFound {
		panic(err)
	}
	acc := foundAcc.Load(w)
	w.accounts[acc.username] = acc
	acc.Login(sock)
	charClients := make([]interface{}, len(acc.chars))
	for i, char := range acc.chars {
		charClients[i] = char.CharClient()
	}
	param := map[string]interface{}{
		"username":    username,
		"charConfigs": charClients,
	}
	clientCall := &ClientCall{
		Receiver: "world",
		Method:   "handleSuccessLoginAcccount",
		Params:   []interface{}{param},
	}
	sock.SendClientCall(clientCall)
	w.logger.Println("Account:", acc.username, "Logined.")
}

func (w *World) LoginAccount(username string, password string, sock *wsConn) {
	_, isOnlineAccount := w.accounts[username]
	if isOnlineAccount {
		clientErr := []interface{}{"wrong username or password"}
		clientCall := &ClientCall{
			Receiver: "world",
			Method:   "handleErrorLoginAccount",
			Params:   clientErr,
		}
		sock.SendClientCall(clientCall)
		return
	}
	foundAcc := &AccountDumpDB{}
	queryAcc := bson.M{"username": username}
	err := w.db.accounts.Find(queryAcc).One(foundAcc)
	if err != nil && err != mgo.ErrNotFound {
		panic(err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(foundAcc.Password), []byte(password))
	if err == mgo.ErrNotFound || err != nil {
		clientErr := []interface{}{"wrong username or password"}
		clientCall := &ClientCall{
			Receiver: "world",
			Method:   "handleErrorLoginAccount",
			Params:   clientErr,
		}
		sock.SendClientCall(clientCall)
		return
	}
	delete(w.accountLoginBySessionMap, username)
	acc := foundAcc.Load(w)
	w.accounts[acc.username] = acc
	acc.Login(sock)
	charClients := make([]interface{}, len(acc.chars))
	for i, char := range acc.chars {
		charClients[i] = char.CharClient()
	}
	param := map[string]interface{}{
		"username":    username,
		"charConfigs": charClients,
	}
	clientCall := &ClientCall{
		Receiver: "world",
		Method:   "handleSuccessLoginAcccount",
		Params:   []interface{}{param},
	}
	sock.SendClientCall(clientCall)
	w.logger.Println("Account:", acc.username, "Logined.")
}

func (w *World) DaoConfigs() *DaoConfigs {
	return w.configs
}

func (w *World) LoadUseSelfFuncByBaseId(baseId int) func(b Bioer) {
	useFunc := w.cache.UseSelfFuncs[baseId]
	if useFunc != nil {
		return useFunc
	}
	item, err := w.NewItemByBaseId(baseId)
	if err != nil || item.ItemTypeByBaseId() != "useSelfItem" {
		return nil
	}
	uItem := item.(*UseSelfItem)
	w.cache.UseSelfFuncs[baseId] = uItem.onUse
	return uItem.onUse
}

func (w *World) ParseUseSelfFuncArrays(useCalls []*UseSelfItemCall, item Itemer) func(b Bioer) {
	// TODO
	// should gen funtion call and put it to return.
	return func(bio Bioer) {
		if bio == nil ||
			reflect.ValueOf(bio).IsNil() {
			return
		}
		for _, uCall := range useCalls {
			_, err := uCall.Eval(item, bio)
			if err != nil {
				w.logger.Println(err)
			}
		}
	}
}

func (w *World) FindSceneByName(name string) *Scene {
	s, ok := w.scenes[name]
	if !ok {
		return nil
	}
	return s
}

func (w *World) NewItemByBaseId(id int) (item Itemer, err error) {
	if id <= 0 {
		return nil, errors.New("out range")
	}
	var iType string
	if id >= 1 && id <= 5000 {
		iType = "equipment"
	} else if id >= 5001 && id <= 10000 {
		iType = "useSelfItem"
	} else {
		iType = "etcItem"
	}
	eqDB := NewEquipment().DB()
	useDump := NewUseSelfItem().DumpDB()
	etcDump := NewEtcItem().DumpDB()
	queryItem := bson.M{"item.baseId": id}
	switch iType {
	case "equipment":
		err = w.db.items.Find(queryItem).One(eqDB)
	case "useSelfItem":
		err = w.db.items.Find(queryItem).One(useDump)
	case "etcItem":
		err = w.db.items.Find(queryItem).One(etcDump)
	}
	if err == nil {
	} else if err != nil && err != mgo.ErrNotFound {
		return
	}
	if err == mgo.ErrNotFound {
		return
	}
	switch iType {
	case "equipment":
		item = eqDB.DumpDB().Load()
	case "useSelfItem":
		config := w.DaoConfigs().ItemConfigs
		useDump.MaxStackCount = config.UseSelfItemConfigs.MaxStackCount
		item = useDump.Load()
		onUse := w.ParseUseSelfFuncArrays(useDump.UseSelfFuncArrays, item)
		item.(*UseSelfItem).onUse = onUse
	case "etcItem":
		config := w.DaoConfigs().ItemConfigs
		etcDump.MaxStackCount = config.EtcItemConfigs.MaxStackCount
		item = etcDump.Load()
	}
	if item.IconViewId() == 0 {
		item.SetIconViewId(item.BaseId())
	}
	if item.BuyPrice() != 0 && item.SellPrice() == 0 {
		item.SetSellPrice(int(float32(item.BuyPrice()) * 0.5))
	}
	return
}

func (w *World) Scenes() map[string]*Scene {
	return w.scenes
}

func (w *World) ScenesSlice() []*Scene {
	ss := make([]*Scene, len(w.scenes))
	i := 0
	for _, s := range w.scenes {
		ss[i] = s
		i++
	}
	return ss
}

func (w *World) Accounts() map[string]*Account {
	return w.accounts
}

func (w *World) Util() *Util {
	return w.util
}

func (w *World) NewWorldTimer(f interface{}, delay time.Duration, interval bool) *WorldTimer {
	fn := reflect.ValueOf(f)
	if fn.Kind() != reflect.Func ||
		fn.Type().NumIn() > 0 {
		return nil
	}
	if 0 >= delay {
		delay = 1
	}
	timer := &WorldTimer{
		duration: delay,
		call:     fn,
		interval: interval,
	}
	w.timers[timer] = timer
	timer.timer = time.AfterFunc(delay, func() {
		w.worldTimer <- timer
	})
	return timer
}

func (w *World) SetInterval(f interface{}, delay time.Duration) *WorldTimer {
	timer := w.NewWorldTimer(f, delay, true)
	if timer == nil {
		return nil
	}
	return timer
}

func (w *World) SetTimeout(f interface{}, delay time.Duration) *WorldTimer {
	timer := w.NewWorldTimer(f, delay, false)
	if timer == nil {
		return nil
	}
	return timer
}

func (w *World) ClearTimeout(timer *WorldTimer) {
	timer.timer.Stop()
	delete(w.timers, timer)
}

func (w *World) ClearInterval(timer *WorldTimer) {
	w.ClearTimeout(timer)
}

func (w *World) TimerEval(timer *WorldTimer) {
	timer.call.Call(nil)
	if timer.interval {
		timer.timer.Reset(timer.duration)
	} else {
		delete(w.timers, timer)
	}
}
