package dao

import (
	"labix.org/v2/mgo"
)

type DaoDB struct {
	url      string
	dbName   string
	session  *mgo.Session
	db       *mgo.Database
	accounts *mgo.Collection
	chars    *mgo.Collection
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
		chars:    db.C("chars"),
		items:    db.C("items"),
	}
	return daoDB, nil
}

func (d *DaoDB) clone() *DaoDB {
	d2 := &DaoDB{}
	d2.url = d.url
	d2.dbName = d.dbName
	d2.session = d.session.Clone()
	d2.db = d2.session.DB(d.dbName)
	d2.accounts = d2.db.C("accounts")
	d2.chars = d2.db.C("chars")
	d2.items = d2.db.C("items")
	return d2
}
