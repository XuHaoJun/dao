package dao

import (
	"encoding/json"
	"gopkg.in/mgo.v2"
	"io/ioutil"
)

type DaoDB struct {
	url      string
	dbName   string
	session  *mgo.Session
	db       *mgo.Database
	accounts *mgo.Collection
	items    *mgo.Collection
}

func NewDaoDB(mgourl string, dbname string) (*DaoDB, error) {
	mongoSession, err := mgo.Dial(mgourl)
	if err != nil {
		return nil, err
	}
	db := mongoSession.DB(dbname)
	daoDB := &DaoDB{
		url:      mgourl,
		dbName:   dbname,
		session:  mongoSession,
		db:       db,
		accounts: db.C("accounts"),
		items:    db.C("items"),
	}
	return daoDB, nil
}

func (d *DaoDB) UpdateAccountIndex() {
	d.accounts.EnsureIndexKey("username")
}

func (d *DaoDB) ImportDefaultJsonDB() error {
	dat, err := ioutil.ReadFile("db/item_db.json")
	var items []interface{}
	err = json.Unmarshal(dat, &items)
	if err != nil {
		return err
	}
	d.items.DropCollection()
	d.items.Insert(items...)
	d.items.EnsureIndexKey("item.baseId")
	return nil
}

func (d *DaoDB) CloneSession() *DaoDB {
	session := d.session.Clone()
	db := session.DB(d.dbName)
	d2 := &DaoDB{
		url:      d.url,
		dbName:   d.dbName,
		session:  session,
		db:       db,
		accounts: db.C("accounts"),
		items:    db.C("items"),
	}
	return d2
}

func (d *DaoDB) Close() {
	d.session.Close()
}
