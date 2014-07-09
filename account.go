package dao

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type Account struct {
	bsonId    bson.ObjectId
	username  string
	password  string
	world     *World
	chars     []*Char
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
	if _, err := accs.UpsertId(a.bsonId, a.DumpDB()); err != nil {
		panic(err)
	}
}

func (a *Account) DoSave() {
	accs := a.db.accounts
	if _, err := accs.UpsertId(a.bsonId, a.DumpDB()); err != nil {
		panic(err)
	}
}

func (a *Account) Save() {
	a.job <- func() {
		a.DoSave()
	}
}

func (a *Account) DumpDB() *AccountDumpDB {
	chars := make([]*CharDumpDB, len(a.chars))
	ci := 0
	for _, char := range a.chars {
		chars[ci] = char.DumpDB()
		ci++
	}
	return &AccountDumpDB{
		Id:       a.bsonId,
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
		checkRange := charSlot >= 0 && charSlot < len(a.chars)
		if a.isOnline == false ||
			checkRange ||
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
		foundChar := struct{ Name string }{}
		queryChar := bson.M{"chars": bson.M{"$elemMatch": bson.M{"name": name}}}
		selectChar := bson.M{"name": 1}
		err := a.db.accounts.Find(queryChar).Select(selectChar).One(&foundChar)
		if err != nil && err != mgo.ErrNotFound {
			panic(err)
		} else if foundChar.Name == name {
			// TODO
			// reject to register same char name
			// send some message to client
			return
		} else if err == mgo.ErrNotFound {
			char := NewChar(name, a)
			char.slotIndex = len(a.chars)
			a.chars = append(a.chars, char)
			char.DoSaveByAccountDB()
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
