package dao

import (
	"errors"

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
	job       chan func()
	quit      chan struct{}
}

type AccountDumpDB struct {
	Id       bson.ObjectId   `bson:"_id"`
	Username string          `bson:"username"`
	Password string          `bson:"password"`
	IsOnline bool            `bson:"isOnline"`
	Chars    []bson.ObjectId `bson:"chars"`
}

func NewAccount(username string, password string, w *World) *Account {
	a := &Account{
		id:       bson.NewObjectId(),
		username: username,
		password: password,
		world:    w,
		isOnline: false,
		db:       w.db.clone(),
		job:      make(chan func(), 16),
		quit:     make(chan struct{}, 1),
	}
	return a
}

func (a *Account) Run() {
	for {
		select {
		case job, ok := <-a.job:
			if !ok {
				return
			}
			job()
		case <-a.quit:
			// may be do some thing..
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
	var chars []bson.ObjectId
	for i, char := range a.chars {
		chars[i] = char.id
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

func (a *Account) Login() {
	go a.Run()
	a.job <- func() {
		a.isOnline = true
		// TODO:
		// update client to selecting char screen
	}
}

func (a *Account) Logout() {
	a.job <- func() {
		a.isOnline = false
		a.usingChar.Save()
		a.usingChar.ShutDown()
		a.ShutDown()
	}
}
