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
	Level     int           `bson:"level"`
	Str       int           `bson:"str"`
	Vit       int           `bson:"vit"`
	Wis       int           `bson:"wis"`
	Spi       int           `bson:"spi"`
	LastScene *SceneInfo    `bson:"lastScene"`
}

func (cDump *CharDumpDB) Load(acc *Account) *Char {
	c := NewChar(cDump.Name, acc)
	c.id = cDump.Id
	c.str = cDump.Str
	c.vit = cDump.Vit
	c.wis = cDump.Wis
	c.spi = cDump.Spi
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
	c.level = 1
	c.str = 1
	c.vit = 1
	c.wis = 1
	c.spi = 1
	return c
}

func (c *Char) Run() {
	c.db = c.account.world.DB().CloneSession()
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

func (c *Char) DoSaveByAccountDB() {
	chars := c.account.db.chars
	if _, err := chars.UpsertId(c.id, c.DumpDB()); err != nil {
		panic(err)
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
	cDump := &CharDumpDB{
		Id:        c.id,
		AccountId: c.account.id,
		Name:      c.name,
		Level:     c.level,
		Str:       c.str,
		Vit:       c.vit,
		Wis:       c.wis,
		Spi:       c.spi,
		LastScene: nil,
	}
	if c.scene != nil {
		cDump.LastScene = &SceneInfo{c.scene.name, c.pos.x, c.pos.y}
	} else {
		cDump.LastScene = &SceneInfo{"daoCity", 0.0, 0.0}
	}
	return cDump
}

func (c *Char) Logout() {
	c.job <- func() {
		c.account.Logout()
	}
}
