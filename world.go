package dao

import (
	"errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"os"
	"reflect"
	"sync"
	"time"
)

type World struct {
	name     string
	server   *Server
	accounts map[string]*Account
	scenes   map[string]*Scene
	db       *DaoDB
	configs  *DaoConfigs
	logger   *log.Logger
	//
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
	interpreter     *WorldInterpreter
	InterpreterEval chan string
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
	RegisterAccount(username string, password string, sock *wsConn)
	LoginAccount(username string, password string, sock *wsConn)
}

func NewWorldByConfig(dc *DaoConfigs) (w *World, err error) {
	w, err = NewWorld(dc.WorldConfigs.Name,
		dc.MongoDBConfigs.URL, dc.MongoDBConfigs.DBName)
	w.configs = dc
	return
}

func NewWorld(name string, mgourl string, dbname string) (*World, error) {
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
		SceneObjecterChangeScene: make(chan *ChangeScene, 256),
		ParseClientCall:          make(chan *WorldParseClientCall, 10240),
		InterpreterEval:          make(chan string, 256),
		BioReborn:                make(chan Bioer, 10240),
		//
		delta:    1.0 / 60.0,
		timeStep: (1.0 * time.Second / 60.0),
		//
		Quit: make(chan struct{}),
		//
		util:  &Util{},
		cache: NewCache(),
	}
	// scenes
	daoCity := NewWallScene(w, "daoCity", 2000, 2000)
	senderNpc := NewNpcByBaseId(w, 1)
	senderNpc.SetPosition(160, 160)
	daoCity.Add(senderNpc.SceneObjecter())
	jackNpc := NewNpcByBaseId(w, 2)
	jackNpc.SetPosition(300, 100)
	daoCity.Add(jackNpc.SceneObjecter())
	w.scenes[daoCity.name] = daoCity
	//
	daoField01 := NewWallScene(w, "daoField01", 6000, 6000)
	daoField01.defaultGroundTextureName = "dirt"
	w.scenes["daoField01"] = daoField01
	// mobs
	var foundScene *Scene
	paul := NewMobByBaseId(w, 1)
	paul.SetPosition(350, 350)
	foundScene = w.FindSceneByName(paul.initSceneName)
	if foundScene != nil {
		foundScene.Add(paul.SceneObjecter())
	}
	// interpreter
	w.interpreter = NewWorldInterpreter(w)
	return w, nil
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
	w.logger.Println("Reloaded DaoConfigs!")
	return
}

func (w *World) ReloadAll() {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		w.ReloadJsonDB()
		wg.Done()
	}()
	go func() {
		w.ReloadDaoConfigs()
		wg.Done()
	}()
	wg.Wait()
}

func (w *World) WorldClientCall() WorldClientCall {
	return w
}

func (w *World) Run() {
	defer w.db.session.Close()
	var wg sync.WaitGroup
	go w.interpreter.ReadRun()
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
		case expr := <-w.InterpreterEval:
			w.interpreter.Eval(expr)
		case acc := <-w.LogoutAccount:
			w.DoLogoutAccount(acc)
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
func (w *World) registerAccount(username string, password string, sock *wsConn) {
	db := w.db.CloneSession()
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
		acc := NewAccount(username, password, w)
		acc.Save()
		go w.db.UpdateAccountIndex()
		// TODO
		// Update client screen
		clientParams := []interface{}{"success register a new account!"}
		clientCall := &ClientCall{
			Receiver: "world",
			Method:   "handleSuccessRegisterAccount",
			Params:   clientParams,
		}
		sock.SendClientCall(clientCall)
	}
}

func (w *World) RegisterAccount(username string, password string, sock *wsConn) {
	go w.registerAccount(username, password, sock)
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

func (w *World) LoginAccount(username string, password string, sock *wsConn) {
	_, ok := w.accounts[username]
	if ok {
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
	if err == mgo.ErrNotFound || foundAcc.Password != password {
		clientErr := []interface{}{"wrong username or password"}
		clientCall := &ClientCall{
			Receiver: "world",
			Method:   "handleErrorLoginAccount",
			Params:   clientErr,
		}
		sock.SendClientCall(clientCall)
		return
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
	eqDump := NewEquipment().DumpDB()
	useDump := NewUseSelfItem().DumpDB()
	etcDump := NewEtcItem().DumpDB()
	queryItem := bson.M{"item.baseId": id}
	switch iType {
	case "equipment":
		err = w.db.items.Find(queryItem).One(eqDump)
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
		item = eqDump.Load()
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
