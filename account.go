package dao

import (
	"errors"
	"strconv"

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

type AccountDumpDB struct {
	Id       bson.ObjectId            `bson:"_id"`
	Username string                   `bson:"username"`
	Password string                   `bson:"password"`
	IsOnline bool                     `bson:"isOnline"`
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
		job:      make(chan func(), 16),
		quit:     make(chan struct{}, 1),
	}
	return a
}

func (a *Account) Run() {
	a.db = a.world.db.CloneSession()
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

func (a *Account) Save() {
	a.job <- func() {
		accs := a.world.db.accounts
		if _, err := accs.UpsertId(a.id, a.DumpDB); err != nil {
			panic(err)
		}
	}
}

func (a *Account) DumpDB() *AccountDumpDB {
	var chars map[string]bson.ObjectId
	for i, char := range a.chars {
		chars[strconv.Itoa(i)] = char.id
	}
	return &AccountDumpDB{
		Id:       a.id,
		Username: a.username,
		Password: a.password,
		IsOnline: a.isOnline,
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

// FIXME
func (a *Account) SelectChar(charPos int) error {
	errC := make(chan error, 1)
	a.job <- func() {
		if a.chars[charPos] == nil {
			errC <- errors.New("Not have char")
		} else {
			a.usingChar = a.chars[charPos]
			close(errC)
		}
	}
	err, ok := <-errC
	if !ok {
		return nil
	}
	return err
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

func (a *Account) Logout() {
	a.job <- func() {
		a.isOnline = false
		a.Save()
		if a.usingChar != nil {
			a.usingChar.Save()
      a.usingChar.ShutDown()
		}
		a.ShutDown()
		// TODO
		// 1. update client to selecting char screen.
		// 2. close socket connection close(a.sock.send)
	}
}
