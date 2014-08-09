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
		job:      make(chan func(), 16),
		quit:     make(chan struct{}),
	}
	return a
}

func (a *Account) AccountClientCall() AccountClientCall {
	return a
}

func (a *Account) DoJob(job func()) (err error) {
	defer handleErrSendCloseChanel(&err)
	a.job <- job
	return
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
			a.DoLogout()
			a.quit <- struct{}{}
			return
		}
	}
}

func (a *Account) ShutDown() <-chan struct{} {
	a.quit <- struct{}{}
	return a.quit
}

func (a *Account) DB() *DaoDB {
	dbC := make(chan *DaoDB, 1)
	err := a.DoJob(func() {
		dbC <- a.db
	})
	if err != nil {
		close(dbC)
		return nil
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
	a.DoJob(func() {
		a.DoSave()
	})
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

func (a *Account) IsSelectingChar() bool {
	c := make(chan bool, 1)
	err := a.DoJob(func() {
		if a.isOnline && a.usingChar == nil {
			c <- true
		} else {
			c <- false
		}
	})
	if err != nil {
		close(c)
		return false
	}
	return <-c
}

func (a *Account) LoginChar(charSlot int) {
	a.DoJob(func() {
		checkRange := charSlot >= 0 && charSlot < len(a.chars)
		if len(a.chars) == 0 ||
			a.isOnline == false ||
			checkRange == false ||
			a.usingChar != nil {
			return
		}
		a.usingChar = a.chars[charSlot]
		// TODO
		// response client to load char's scene
		clientCall := &ClientCall{
			Receiver: "account",
			Method:   "handleSuccessLoginChar",
			Params:   nil,
		}
		a.sock.SendMsg(clientCall)
		a.usingChar.Login()
	})
}

func (a *Account) Login(sock *wsConn) {
	go a.Run()
	a.DoJob(func() {
		if a.isOnline == true {
			return
		}
		a.isOnline = true
		a.sock = sock
	})
}

func (a *Account) UsingChar() *Char {
	cC := make(chan *Char, 1)
	err := a.DoJob(func() {
		cC <- a.usingChar
	})
	if err != nil {
		close(cC)
		return nil
	}
	return <-cC
}

func (a *Account) CreateChar(name string) {
	a.DoJob(func() {
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
		err := a.db.accounts.Find(queryChar).Select(bson.M{"_id": 1}).One(&struct{}{})
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
			char.DoSave()
			a.world.logger.Println(
				"Account:", a.username,
				"created a new char:",
				char.name+".")
			// TODO
			// Update client screen
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
	})
}

func (a *Account) DoLogout() {
	if a.isOnline == false {
		return
	}
	a.isOnline = false
	if a.usingChar != nil {
		a.usingChar.DoSave()
		a.usingChar.ShutDown()
	}
	a.world.LogoutAccount(a.username)
	a.sock.Close()
	a.world.logger.Println("Account:", a.username, "logouted.")
}

func (a *Account) Logout() {
	<-a.ShutDown()
}
