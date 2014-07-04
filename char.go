package dao

import (
	"labix.org/v2/mgo/bson"
)

type Char struct {
	*BattleBioBase
	id        bson.ObjectId
	account   *Account
	db        *DaoDB
	sock      *wsConn
	lastScene *SceneInfo
}

type CharDumpDB struct {
	Id        bson.ObjectId `bson:"_id"`
	AccountId bson.ObjectId `bson:"accountId"`
	Name      string        `bson:"name"`
	LastScene *SceneInfo    `bson:"lastScene"`
}

func (cDump *CharDumpDB) Load(acc *Account) *Char {
	c := NewChar(cDump.Name, acc)
	c.id = cDump.Id
	c.lastScene = cDump.LastScene
	return c
}

func NewChar(name string, acc *Account) *Char {
	c := &Char{
		BattleBioBase: NewBattleBioBase(),
		id:            bson.NewObjectId(),
		account:       acc,
		sock:          acc.sock,
	}
	c.name = name
	return c
}

func (c *Char) Run() {
	c.db = c.account.world.db.CloneSession()
	defer c.db.session.Close()
	for {
		select {
		case job, ok := <-c.job:
			if !ok {
				return
			}
			job()
		case <-c.quit:
			c.quit <- struct{}{}
			return
		}
	}
}

func (c *Char) Save() {
	c.job <- func() {
		chars := c.db.chars
		if _, err := chars.UpsertId(c.id, c.DumpDB()); err != nil {
			panic(err)
		}
	}
}

func (c *Char) DumpDB() *CharDumpDB {
	return &CharDumpDB{
		Id:        c.id,
		AccountId: c.account.id,
		Name:      c.name,
		LastScene: &SceneInfo{c.scene.name, c.pos.x, c.pos.y},
	}
}

func (c *Char) Logout() {
	c.job <- func() {
		c.account.Logout()
	}
}
