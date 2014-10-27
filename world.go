package dao

import (
	"encoding/json"
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
	accounts map[string]*Account
	scenes   map[string]*Scene
	db       *DaoDB
	configs  *WorldConfigs
	logger   *log.Logger
	//
	// SaveAccount     chan *Account
	// RegisterAccount chan *WorldRegisterAccount
	// LoginAccount    chan *WorldLoginAccount
	LogoutAccount chan *Account
	//
	SceneObjecterChangeScene chan *ChangeScene
	//
	// AccountLoginChar  chan *AccountLoginChar
	// AccountCreateChar chan *AccountCreateChar
	//
	ParseClientCall chan *WorldParseClientCall
	//
	delta    float32
	timeStep time.Duration
	//
	job  chan func()
	Quit chan struct{}
}

type ChangeScene struct {
	SceneObjecter SceneObjecter
	Scene         *Scene
}

type WorldConfigs struct {
	maxChars     int
	maxCharItems int
}

type WorldClientCall interface {
	RegisterAccount(username string, password string)
	LoginAccount(username string, password string, sock *wsConn)
}

func NewWorld(name string, mgourl string, dbname string) (*World, error) {
	db, err := NewDaoDB(mgourl, dbname)
	if err != nil {
		return nil, err
	}
	err = db.ImportDefaultJsonDB()
	if err != nil {
		return nil, err
	}
	w := &World{
		name:                     name,
		accounts:                 make(map[string]*Account),
		scenes:                   make(map[string]*Scene),
		db:                       db,
		configs:                  &WorldConfigs{5, 30},
		logger:                   log.New(os.Stdout, "[dao-"+name+"] ", 0),
		LogoutAccount:            make(chan *Account, 8),
		SceneObjecterChangeScene: make(chan *ChangeScene, 256),
		ParseClientCall:          make(chan *WorldParseClientCall, 10240),
		//
		delta:    1.0 / 15.0,
		timeStep: (1.0 * time.Second / 15.0),
		//
		Quit: make(chan struct{}),
	}
	baseScene := NewWallScene(w, "daoCity", 2000, 2000)
	senderNpc := NewNpcByBaseId(w, 1)
	baseScene.Add(senderNpc.SceneObjecter())
	jackNpc := NewNpcByBaseId(w, 2)
	jackNpc.SetPosition(300, 100)
	baseScene.Add(jackNpc.SceneObjecter())
	w.scenes[baseScene.name] = baseScene
	return w, nil
}

func (w *World) WorldClientCall() WorldClientCall {
	return w
}

func (w *World) Run() {
	defer w.db.session.Close()
	var wg sync.WaitGroup
	physicC := time.Tick(w.timeStep)
	for {
		select {
		case acc := <-w.LogoutAccount:
			w.DoLogoutAccount(acc)
		case params := <-w.ParseClientCall:
			w.DoParseClientCall(params.Msg, params.Conn)
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
		case <-physicC:
			for _, scene := range w.scenes {
				wg.Add(1)
				go func(s *Scene, dt float32) {
					s.Update(dt)
					wg.Done()
				}(scene, w.delta)
			}
			wg.Wait()
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
func (w *World) registerAccount(username string, password string) {
	db := w.db.CloneSession()
	queryAcc := bson.M{"username": username}
	err := db.accounts.Find(queryAcc).Select(bson.M{"_id": 1}).One(&struct{}{})
	if err != nil && err != mgo.ErrNotFound {
		panic(err)
	} else if err != mgo.ErrNotFound {
		// TODO
		// reject to register same user
		// send some message to client
	} else if err == mgo.ErrNotFound {
		acc := NewAccount(username, password, w)
		acc.Save()
		w.logger.Println("Account:", acc.username, "Registered.")
		// TODO
		// Update client screen
	}
}

func (w *World) RegisterAccount(username string, password string) {
	go w.registerAccount(username, password)
}

type WorldParseClientCall struct {
	Msg  []byte
	Conn *wsConn
}

func (w *World) DoParseClientCall(msg []byte, conn *wsConn) {
	logger := w.logger
	acc := conn.account
	clientCall := &ClientCall{}
	err := json.Unmarshal(msg, clientCall)
	if err != nil {
		logger.Println(conn.ws.RemoteAddr(), ": can't parse to json:", string(msg))
		return
	}
	// logger.Println(conn.ws.RemoteAddr(), "call:", clientCall)
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
		if clientCall.Method == "LoginAccount" {
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

func (w *World) RequestParseClientCall(msg []byte, conn *wsConn) {
	w.ParseClientCall <- &WorldParseClientCall{msg, conn}
}

func (w *World) DoLogoutAccount(acc *Account) {
	_, ok := w.accounts[acc.username]
	if ok {
		acc.Logout()
		delete(w.accounts, acc.username)
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
		sock.SendMsg(clientCall)
		return
	}
	foundAcc := &AccountDumpDB{}
	queryAcc := bson.M{"username": username, "password": password}
	err := w.db.accounts.Find(queryAcc).One(foundAcc)
	if err != nil && err != mgo.ErrNotFound {
		panic(err)
	}
	if err == mgo.ErrNotFound {
		clientErr := []interface{}{"wrong username or password"}
		clientCall := &ClientCall{
			Receiver: "world",
			Method:   "handleErrorLoginAccount",
			Params:   clientErr,
		}
		sock.SendMsg(clientCall)
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
	sock.SendMsg(clientCall)
	w.logger.Println("Account:", acc.username, "Logined.")
}

func (w *World) Configs() *WorldConfigs {
	return w.configs
}

func (w *World) NewEquipmentByBaseId(id int) (*Equipment, error) {
	eqDump := NewEquipment().DumpDB()
	// TODO
	// add check db error
	err := w.db.items.Find(bson.M{"item.baseId": id}).One(eqDump)
	if err != nil {
		return nil, err
	}
	eq := eqDump.Load()
	if eq.iconViewId == 0 {
		eq.iconViewId = eq.baseId
	}
	return eq, nil
}

// func (w *World) findItemByBaseId(iType string, id int, iDump interface{}) (err error) {
// 	queryItem := bson.M{iType: bson.M{"$elemMatch": bson.M{"item.baseId": id}}}
// 	switch iDumpDB := iDump.(type) {
// 	case *EquipmentDumpDB:
// 		err = w.db.items.Find(queryItem).One(iDumpDB)
// 	case *UseSelfItemDumpDB:
// 		err = w.db.items.Find(queryItem).One(iDumpDB)
// 	case *EtcItemDumpDB:
// 		err = w.db.items.Find(queryItem).One(iDumpDB)
// 	}
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

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
	w.logger.Println(iType)
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
	w.logger.Println(err)
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
		item = useDump.Load()
	case "etcItem":
		item = etcDump.Load()
	}
	if item.IconViewId() == 0 {
		item.SetIconViewId(item.BaseId())
	}
	if item.BuyPrice() != 0 && item.SellPrice() == 0 {
		item.SetSellPrice(int(float32(item.BuyPrice()) * 0.5))
	}
	w.logger.Println(item)
	w.logger.Println(item.Name())
	w.logger.Println(item.ItemTypeByBaseId())
	return
}
