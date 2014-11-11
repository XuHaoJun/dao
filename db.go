package dao

import (
	"gopkg.in/mgo.v2"
	"os/exec"
	"sync"
)

type DaoDB struct {
	url         string
	dbName      string
	session     *mgo.Session
	db          *mgo.Database
	accounts    *mgo.Collection
	items       *mgo.Collection
	updateMutex *sync.Mutex
}

func NewDaoDB(mgourl string, dbname string) (*DaoDB, error) {
	mongoSession, err := mgo.Dial(mgourl)
	if err != nil {
		return nil, err
	}
	db := mongoSession.DB(dbname)
	daoDB := &DaoDB{
		url:         mgourl,
		dbName:      dbname,
		session:     mongoSession,
		db:          db,
		accounts:    db.C("accounts"),
		items:       db.C("items"),
		updateMutex: &sync.Mutex{},
	}
	return daoDB, nil
}

func (d *DaoDB) UpdateAccountIndex() {
	d.updateMutex.Lock()
	b := d.CloneSession()
	b.accounts.EnsureIndexKey("username")
	d.updateMutex.Unlock()
}

func (d *DaoDB) ImportDefaultJsonDB() error {
	itemCmd := exec.Command("mongoimport", "--db", d.dbName,
		"--collection", "items", "--type", "json",
		"--file", "db/item_db.json", "--quiet", "--jsonArray",
		"--upsert", "--upsertFields", "item.baseId")
	err := itemCmd.Run()
	if err != nil {
		return err
	}
	d.items.EnsureIndexKey("item.baseId")
	return nil
}

func (d *DaoDB) CloneSession() *DaoDB {
	session := d.session.Clone()
	db := session.DB(d.dbName)
	d2 := &DaoDB{
		url:         d.url,
		dbName:      d.dbName,
		session:     session,
		db:          db,
		accounts:    db.C("accounts"),
		items:       db.C("items"),
		updateMutex: d.updateMutex,
	}
	return d2
}
