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
	}
	return daoDB, nil
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
	}
	return d2
}
