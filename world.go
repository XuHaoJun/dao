package dao

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type World struct {
	name     string
	accounts map[string]*Account
	scenes   map[string]*Scene
	db       *DaoDB
	job      chan func()
	quit     chan struct{}
}

type WorldClientCall interface {
	RegisterAccount(username string, password string)
	LoginAccount(username string, password string, sock *wsConn) *Account
	LogoutAccount(username string)
}

func NewWorld(name string, mgourl string, dbname string) (*World, error) {
	db, err := NewDaoDB(mgourl, dbname)
	if err != nil {
		panic(err)
	}
	w := &World{
		name:     name,
		db:       db,
		accounts: make(map[string]*Account),
		scenes:   make(map[string]*Scene),
		job:      make(chan func(), 512),
		quit:     make(chan struct{}, 1),
	}
	return w, nil
}

func (w *World) WorldClientCall() WorldClientCall {
	return w
}

func (w *World) Run() {
	defer w.db.session.Close()
	for {
		select {
		case job, ok := <-w.job:
			if !ok {
				return
			}
			job()
		case <-w.quit:
			for _, acc := range w.accounts {
				// may put acc save to shutdown
				acc.Save()
				if acc.usingChar != nil {
					acc.usingChar.Save()
				}
				acc.ShutDown()
				acc.usingChar.ShutDown()
			}
			w.quit <- struct{}{}
			return
		}
	}
}

func (w *World) ShutDown() {
	w.quit <- struct{}{}
	<-w.quit
}

func (w *World) RegisterAccount(username string, password string) {
	w.job <- func() {
		foundAcc := AccountDumpDB{}
		err := w.db.accounts.Find(bson.M{"username": username}).One(&foundAcc)
		if err != nil && err != mgo.ErrNotFound {
			// should return error?
			return
		}
		if foundAcc.Username == username {
			// TODO
			// reject to register same user
			// send some message to client
			return
		}
		acc := NewAccount(username, password, w)
		acc.DoSave()
	}
}

func (w *World) LoginAccount(username string, password string, sock *wsConn) *Account {
	accC := make(chan *Account, 1)
	w.job <- func() {
		_, ok := w.accounts[username]
		if ok {
			close(accC)
			return
		}
		foundAcc := &AccountDumpDB{}
		queryAcc := bson.M{"username": username, "password": password}
		err := w.db.accounts.Find(queryAcc).One(foundAcc)
		if err != nil {
			panic(err)
		}
		if foundAcc.Username == "" {
			// notify client not find or password error
			close(accC)
			return
		}
		acc := foundAcc.Load(w)
		w.accounts[acc.username] = acc
		acc.Login(sock)
		accC <- acc
	}
	acc, ok := <-accC
	if !ok {
		return nil
	}
	return acc
}

func (w *World) LogoutAccount(username string) {
	w.job <- func() {
		acc, ok := w.accounts[username]
		if !ok {
			// should return error
		} else {
			delete(w.accounts, username)
			acc.Logout()
		}
	}
}

// func (w *World) IsOnlineAccountByUsername(username string) bool {
// 	has := make(chan bool, 1)
// 	w.job <- func() {
// 		_, ok := w.accounts[username]
// 		has <- ok
// 	}
// 	return <-has
// }

// func (w *World) IsOnlineAccount(acc *Account) bool {
// 	has := make(chan bool, 1)
// 	w.job <- func() {
// 		_, ok := w.accounts[acc]
// 		has <- ok
// 	}
// 	return <-has
// }

func (w *World) FindSceneByName(sname string) *Scene {
	sceneChan := make(chan *Scene, 1)
	w.job <- func() {
		sceneChan <- w.scenes[sname]
	}
	scene, ok := <-sceneChan
	if !ok {
		return nil
	}
	return scene
}

func (w *World) AddScene(s *Scene) {
	w.job <- func() {
		w.scenes[s.name] = s
		// may be active scene, like s.Run()
	}
}
