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
	configs  *WorldConfigs
	job      chan func()
	quit     chan struct{}
}

type WorldConfigs struct {
	maxCharItems int
}

type WorldClientCall interface {
	RegisterAccount(username string, password string)
	LoginAccount(username string, password string, sock *wsConn) *Account
}

func NewWorld(name string, mgourl string, dbname string) (*World, error) {
	db, err := NewDaoDB(mgourl, dbname)
	baseScene := NewScene("daoCity")
	if err != nil {
		panic(err)
	}
	w := &World{
		name:     name,
		accounts: make(map[string]*Account),
		scenes:   make(map[string]*Scene),
		db:       db,
		configs:  &WorldConfigs{40},
		job:      make(chan func(), 512),
		quit:     make(chan struct{}, 1),
	}
	w.scenes[baseScene.name] = baseScene
	return w, nil
}

func (w *World) WorldClientCall() WorldClientCall {
	return w
}

func (w *World) Run() {
	defer w.db.session.Close()
	go w.scenes["daoCity"].Run()
	for {
		select {
		case job, ok := <-w.job:
			if !ok {
				return
			}
			job()
		case <-w.quit:
			for _, acc := range w.accounts {
				acc.Logout()
			}
			w.quit <- struct{}{}
			return
		}
	}
}

func (w *World) DB() *DaoDB {
	dbC := make(chan *DaoDB, 1)
	w.job <- func() {
		dbC <- w.db
	}
	return <-dbC
}

func (w *World) ShutDown() {
	w.quit <- struct{}{}
	<-w.quit
}

func (w *World) RegisterAccount(username string, password string) {
	w.job <- func() {
		foundAcc := struct {
			Username string `bson:"username"`
		}{}
		queryAcc := bson.M{"username": username}
		selectAcc := bson.M{"username": 1}
		err := w.db.accounts.Find(queryAcc).Select(selectAcc).One(&foundAcc)
		if err != nil && err != mgo.ErrNotFound {
			panic(err)
		} else if foundAcc.Username == username {
			// TODO
			// reject to register same user
			// send some message to client
			return
		} else if err == mgo.ErrNotFound {
			acc := NewAccount(username, password, w)
			acc.DoSaveByWorldDB()
			// TODO
			// Update client screen
		}
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
		if err != nil && err != mgo.ErrNotFound {
			panic(err)
		}
		if err == mgo.ErrNotFound {
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
		_, ok := w.accounts[username]
		if !ok {
			// should return error
		} else {
			delete(w.accounts, username)
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

func (w *World) IsOnlineAccount(acc *Account) bool {
	has := make(chan bool, 1)
	w.job <- func() {
		_, ok := w.accounts[acc.username]
		has <- ok
	}
	return <-has
}

func (w *World) Configs() *WorldConfigs {
	return w.configs
}

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
	}
}

func (w *World) AddSceneAndRun(s *Scene) {
	w.job <- func() {
		w.scenes[s.name] = s
		go s.Run()
	}
}
