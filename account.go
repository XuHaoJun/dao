package dao

import (
	"strconv"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type Account struct {
	id        bson.ObjectId
	username  string
	password  string
	world     *World
	chars     map[int]*Char
	usingChar *Char
	isOnline  bool
	db        *DaoDB
	sock      *wsConn
	job       chan func()
	quit      chan struct{}
}

type AccountClientCall interface {
	CreateChar(name string)
	LoginChar(charSlog int)
	Logout()
}

type AccountDumpDB struct {
	Id       bson.ObjectId            `bson:"_id"`
	Username string                   `bson:"username"`
	Password string                   `bson:"password"`
	Chars    map[string]bson.ObjectId `bson:"chars"`
}

func (aDump *AccountDumpDB) Load(w *World) *Account {
	acc := NewAccount(aDump.Username, aDump.Password, w)
	acc.id = aDump.Id
	for s, cid := range aDump.Chars {
		cDump := &CharDumpDB{}
		err := w.db.chars.FindId(cid).One(cDump)
		if err != nil {
			panic(err)
		}
		char := cDump.Load(acc)
		index, _ := strconv.Atoi(s)
		acc.chars[index] = char
	}
	return acc
}

func NewAccount(username string, password string, w *World) *Account {
	a := &Account{
		id:       bson.NewObjectId(),
		username: username,
		password: password,
		world:    w,
		chars:    make(map[int]*Char),
		isOnline: false,
		job:      make(chan func(), 128),
		quit:     make(chan struct{}, 1),
	}
	return a
}

func (a *Account) AccountClientCall() AccountClientCall {
	return a
}

func (a *Account) Run() {
	a.db = a.world.DB().CloneSession()
	defer a.db.session.Close()
	for {
		select {
		case job, ok := <-a.job:
			if !ok {
				return
			}
			job()
		case <-a.quit:
			a.quit <- struct{}{}
			return
		}
	}
}

func (a *Account) ShutDown() {
	a.quit <- struct{}{}
	<-a.quit
}

func (a *Account) DB() *DaoDB {
	dbC := make(chan *DaoDB, 1)
	a.job <- func() {
		dbC <- a.db
	}
	return <-dbC
}

func (a *Account) DoSaveByWorldDB() {
	accs := a.world.db.accounts
	if _, err := accs.UpsertId(a.id, a.DumpDB()); err != nil {
		panic(err)
	}
}

func (a *Account) DoSave() {
	accs := a.db.accounts
	if _, err := accs.UpsertId(a.id, a.DumpDB()); err != nil {
		panic(err)
	}
}

func (a *Account) Save() {
	a.job <- func() {
		a.DoSave()
	}
}

func (a *Account) DumpDB() *AccountDumpDB {
	chars := make(map[string]bson.ObjectId)
	for i, char := range a.chars {
		chars[strconv.Itoa(i)] = char.id
	}
	return &AccountDumpDB{
		Id:       a.id,
		Username: a.username,
		Password: a.password,
		Chars:    chars,
	}
}

func (a *Account) IsSelectingChar() bool {
	c := make(chan bool, 1)
	a.job <- func() {
		if a.isOnline && a.usingChar == nil {
			c <- true
		} else {
			c <- false
		}
	}
	return <-c
}

func (a *Account) LoginChar(charSlot int) {
	a.job <- func() {
		if a.isOnline == false ||
			a.chars[charSlot] == nil ||
			a.usingChar != nil {
			return
		}
		a.usingChar = a.chars[charSlot]
		a.usingChar.Login()
	}
}

func (a *Account) Login(sock *wsConn) {
	go a.Run()
	a.job <- func() {
		a.isOnline = true
		a.Save()
		a.sock = sock
		// TODO
		// update client to selecting char screen
	}
}

func (a *Account) UsingChar() *Char {
	cC := make(chan *Char, 1)
	a.job <- func() {
		cC <- a.usingChar
	}
	return <-cC
}

func (a *Account) CreateChar(name string) {
	a.job <- func() {
		if a.isOnline == false {
			return
		}
		foundChar := CharDumpDB{}
		err := a.db.chars.Find(bson.M{"name": name}).One(&foundChar)
		if err != nil && err != mgo.ErrNotFound {
			panic(err)
		} else if foundChar.Name == name {
			// TODO
			// reject to register same char name
			// send some message to client
			return
		} else if err == mgo.ErrNotFound {
			char := NewChar(name, a)
			char.DoSaveByAccountDB()
			charsSlotPos := len(a.chars)
			a.chars[charsSlotPos] = char
			a.DoSave()
			// TODO
			// Update client screen
		}
	}
}

func (a *Account) Logout() {
	a.job <- func() {
		if a.isOnline == false {
			return
		}
		if a.usingChar != nil {
			a.usingChar.Logout()
		}
		a.world.LogoutAccount(a.username)
		a.Save()
		a.ShutDown()
		// TODO
		// 1. update client to selecting char screen.
		// 2. close socket connection close(a.sock.send)
	}
}
