package dao

import (
	"strconv"

	"labix.org/v2/mgo/bson"
)

type Char struct {
	*BattleBioBase
	bsonId      bson.ObjectId
	usingEquips *Equips
	items       *Items
	world       *World
	account     *Account
	slotIndex   int
	db          *DaoDB
	isOnline    bool
	sock        *wsConn
	lastScene   *SceneInfo
}

type CharDumpDB struct {
	Id        bson.ObjectId `bson:"_id"`
	SlotIndex int           `bson:"slotIndex"`
	Name      string        `bson:"name"`
	Level     int           `bson:"level"`
	Str       int           `bson:"str"`
	Vit       int           `bson:"vit"`
	Wis       int           `bson:"wis"`
	Spi       int           `bson:"spi"`
	LastScene *SceneInfo    `bson:"lastScene"`
	Items     *ItemsDumpDB  `bson:"items"`
}

type CharClientCall interface {
	Logout()
}

func (cDump *CharDumpDB) Load(acc *Account) *Char {
	c := NewChar(cDump.Name, acc)
	c.slotIndex = cDump.SlotIndex
	c.bsonId = cDump.Id
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
		bsonId:        bson.NewObjectId(),
		usingEquips:   &Equips{},
		items:         NewItems(),
		isOnline:      false,
		account:       acc,
		world:         acc.world,
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
	accs := c.account.db.accounts
	ci := strconv.Itoa(c.slotIndex)
	cii := "chars." + ci
	update := bson.M{"$set": bson.M{cii: c.DumpDB()}}
	if err := accs.UpdateId(c.account.bsonId, update); err != nil {
		panic(err)
	}
}

func (c *Char) DoSave() {
	accs := c.db.accounts
	ci := strconv.Itoa(c.slotIndex)
	cii := "chars." + ci
	update := bson.M{"$set": bson.M{cii: c.DumpDB()}}
	if err := accs.UpdateId(c.account.bsonId, update); err != nil {
		panic(err)
	}
}

func (c *Char) Save() {
	c.job <- func() {
		c.DoSave()
	}
}

func (c *Char) DumpDB() *CharDumpDB {
	cDump := &CharDumpDB{
		Id:        c.bsonId,
		SlotIndex: c.slotIndex,
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

func (c *Char) Login() {
	go c.Run()
	c.job <- func() {
		c.isOnline = true
		scene := c.world.FindSceneByName(c.lastScene.Name)
		if scene == nil {
			// TODO
			// imple saveScene on char
			// c.world.FindSceneByName(c.saveScene.Name)
			return
		}
		scene.AddBio(c)
	}
}

func (c *Char) Logout() {
	c.job <- func() {
		if c.isOnline == false {
			return
		}
		c.Save()
		c.account.Logout()
		c.ShutDown()
	}
}
