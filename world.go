package dao

import (
  "labix.org/v2/mgo/bson"
)

type World struct {
  name string
  accounts map[*Account]struct{}
  scenes map[*Scene]struct{}
  db *DaoDB
  job chan func()
  quit chan struct{}
}

func NewWorld(name string, mgourl string, dbname string) (*World, error) {
  db, err := NewDaoDB(mgourl, dbname)
  if err != nil {
    return nil, err
  }
  w := &World{
    name: name,
    db: db,
    job: make(chan func(), 512),
    quit: make(chan struct{}, 1),
  }
  return w, nil
}

func (w *World) Run() {
  for {
    select {
    case job, ok := <-w.job:
      if !ok {
        return
      }
      job()
    case <-w.quit:
      for acc, _ := range w.accounts {
        // may put acc save to shutdown
        acc.Save()
        if acc.usingChar != nil {
          acc.usingChar.Save()
          acc.usingChar.db.session.Close()
        }
        acc.db.session.Close()
        acc.ShutDown()
        acc.usingChar.ShutDown()
      }
      w.db.session.Close()
      w.quit <-struct{}{}
      return
    }
  }
}

func (w *World) ShutDown() {
  w.quit <-struct{}{}
  <-w.quit
}

func (w *World) RegisterAccount(username string, password string) error {
  w.job <-func() {
    foundAcc := AccountDumpDB{}
    err := w.db.accounts.Find(bson.M{"username": username}).One(foundAcc)
    if err != nil {
      panic(err)
    }
    if foundAcc.Username == username {
      // TODO: reject to register same user
      // return error
    }
    acc := NewAccount(username, password, w)
    acc.Save()
  }
  return nil
}

func (w *World) HasAccount(acc *Account) bool {
  has := make(chan bool, 1)
  w.job <-func() {
    _, ok := w.accounts[acc]
    has <-ok
  }
  return <-has
}

func (w *World) AddAccount(acc *Account) {
  w.job <-func() {
    w.accounts[acc] = struct{}{}
    acc.Login()
  }
}

func (w *World) RemoveAccount(acc *Account) {
  w.job <-func() {
    delete(w.accounts, acc)
  }
}

func (w *World) FindSceneByName(sname string) *Scene {
  sceneChan := make(chan *Scene, 1)
  w.job <-func() {
    var foundScene *Scene
    for scene, _ := range w.scenes {
      if scene.name == sname {
        foundScene = scene
        break
      }
    }
    sceneChan <-foundScene
  }
  return <-sceneChan
}

func (w *World) AddScene(s *Scene) {
  w.job <-func() {
    w.scenes[s] = struct{}{}
  }
}
