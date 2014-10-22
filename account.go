package dao

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Account struct {
	bsonId    bson.ObjectId
	username  string
	password  string
	world     *World
	chars     []*Char
	usingChar *Char
	isOnline  bool
	sock      *wsConn
}

type AccountClientCall interface {
	CreateChar(name string)
	LoginChar(charSlog int)
	Logout()
}

type AccountDumpDB struct {
	Id       bson.ObjectId `bson:"_id"`
	Username string        `bson:"username"`
	Password string        `bson:"password"`
	Chars    []*CharDumpDB `bson:"chars"`
}

func (aDump *AccountDumpDB) Load(w *World) *Account {
	acc := NewAccount(aDump.Username, aDump.Password, w)
	acc.bsonId = aDump.Id
	acc.chars = make([]*Char, len(aDump.Chars))
	for i, charDump := range aDump.Chars {
		acc.chars[i] = charDump.Load(acc)
	}
	return acc
}

func NewAccount(username string, password string, w *World) *Account {
	a := &Account{
		bsonId:   bson.NewObjectId(),
		username: username,
		password: password,
		world:    w,
		chars:    []*Char{},
		isOnline: false,
	}
	return a
}

func (a *Account) AccountClientCall() AccountClientCall {
	return a
}

func (a *Account) Save() {
	accs := a.world.db.accounts
	if _, err := accs.UpsertId(a.bsonId, a.DumpDB()); err != nil {
		panic(err)
	}
}

func (a *Account) DumpDB() *AccountDumpDB {
	chars := make([]*CharDumpDB, len(a.chars))
	for i, char := range a.chars {
		chars[i] = char.DumpDB()
	}
	return &AccountDumpDB{
		Id:       a.bsonId,
		Username: a.username,
		Password: a.password,
		Chars:    chars,
	}
}

func (a *Account) LoginChar(charSlot int) {
	checkRange := charSlot >= 0 && charSlot < len(a.chars)
	if len(a.chars) == 0 ||
		a.isOnline == false ||
		checkRange == false ||
		a.usingChar != nil {
		return
	}
	a.usingChar = a.chars[charSlot]
	a.usingChar.sock = a.sock
	a.usingChar.Login()
	// TODO
	// response client to load char's scene
	clientCalls := make([]*ClientCall, 5)
	scene := a.usingChar.scene
	// FIXME
	// sceneParam := map[string]interface{}{
	// 	"name":   scene.name,
	// 	"width":  600,
	// 	"height": 600,
	// }
	sceneParam := a.usingChar.scene.SceneClient()
	clientCalls[0] = &ClientCall{
		Receiver: "world",
		Method:   "handleAddScene",
		Params:   []interface{}{sceneParam},
	}
	accParam := map[string]interface{}{"usingChar": charSlot}
	clientCalls[1] = &ClientCall{
		Receiver: "account",
		Method:   "handleSuccessLoginChar",
		Params:   []interface{}{accParam},
	}
	clientCalls[2] = &ClientCall{
		Receiver: "world",
		Method:   "handleRunScene",
		Params:   []interface{}{scene.name},
	}
	char := a.usingChar
	charParam := map[string]interface{}{
		"sceneName": scene.name,
		"id":        char.id,
	}
	clientCalls[3] = &ClientCall{
		Receiver: "char",
		Method:   "handleJoinScene",
		Params:   []interface{}{charParam},
	}
	clientCalls[4] = &ClientCall{
		Receiver: "char",
		Method:   "handleChatMessage",
		Params: []interface{}{
			&ChatMessageClient{
				"System",
				"",
				"Welcom to Dao!",
			},
		},
	}
	a.sock.SendMsg(clientCalls)
}

func (a *Account) Login(sock *wsConn) {
	if a.isOnline == true {
		return
	}
	a.isOnline = true
	a.sock = sock
	sock.account = a
}

func (a *Account) UsingChar() *Char {
	return a.usingChar
}

func (a *Account) CreateChar(name string) {
	if a.isOnline == false {
		return
	}
	if len(a.chars) >= a.world.Configs().maxChars {
		clientCall := &ClientCall{
			Receiver: "account",
			Method:   "handleErrorCreateChar",
			Params:   []interface{}{"overflow max chars."},
		}
		a.sock.SendMsg(clientCall)
		return
	}
	queryChar := bson.M{"chars": bson.M{"$elemMatch": bson.M{"name": name}}}
	err := a.world.db.accounts.Find(queryChar).Select(bson.M{"_id": 1}).One(&struct{}{})
	if err != nil && err != mgo.ErrNotFound {
		panic(err)
	} else if err != mgo.ErrNotFound {
		clientCall := &ClientCall{
			Receiver: "account",
			Method:   "handleErrorCreateChar",
			Params:   []interface{}{"duplicate char name."},
		}
		a.sock.SendMsg(clientCall)
	} else if err == mgo.ErrNotFound {
		char := NewChar(name, a)
		char.slotIndex = len(a.chars)
		a.chars = append(a.chars, char)
		char.Save()
		a.world.logger.Println(
			"Account:", a.username,
			"created a new char:",
			char.name+".")
		param := map[string]interface{}{
			"charConfig": char.CharClient(),
		}
		clientCall := &ClientCall{
			Receiver: "account",
			Method:   "handleSuccessCreateChar",
			Params:   []interface{}{param},
		}
		a.sock.SendMsg(clientCall)
	}
}

func (a *Account) Logout() {
	if a.isOnline == false {
		return
	}
	a.isOnline = false
	if a.usingChar != nil {
		c := a.usingChar
		if c.isOnline == false {
			return
		}
		c.isOnline = false
		c.Save()
		if c.scene != nil {
			c.scene.Remove(c)
		}
		c.account.world.logger.Println("Char:", c.name, "logouted.")
	}
	a.sock.Close()
	a.world.logger.Println("Account:", a.username, "logouted.")
}

func (a *Account) RequestLogout() {
	a.world.LogoutAccount <- a
}
