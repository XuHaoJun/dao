package dao

import (
	"gopkg.in/mgo.v2"
	"os/exec"
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
	itemCmd := exec.Command("mongoimport", "--db", d.dbName,
		"--collection", "items", "--type", "json",
		"--file", "db/item_db.json", "--quiet", "--jsonArray",
		"--upsert", "--upsertFields", "item.baseId")
	err := itemCmd.Run()
	if err != nil {
		return err
	}
	go d.items.EnsureIndexKey("item.baseId")
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
