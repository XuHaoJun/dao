package dao

import (
	"strconv"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type CharClientCall interface {
	Logout()
}

type Char struct {
	*BattleBioBase
	bsonId      bson.ObjectId
	usingEquips UsingEquips
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
	Id          bson.ObjectId     `bson:"_id"`
	SlotIndex   int               `bson:"slotIndex"`
	Name        string            `bson:"name"`
	Level       int               `bson:"level"`
	Str         int               `bson:"str"`
	Vit         int               `bson:"vit"`
	Wis         int               `bson:"wis"`
	Spi         int               `bson:"spi"`
	LastScene   *SceneInfo        `bson:"lastScene"`
	UsingEquips UsingEquipsDumpDB `bson:"usingEquips"`
	Items       *ItemsDumpDB      `bson:"items"`
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
	c.items = cDump.Items.Load()
	c.usingEquips = cDump.UsingEquips.Load()
	return c
}

func NewChar(name string, acc *Account) *Char {
	c := &Char{
		BattleBioBase: NewBattleBioBase(),
		bsonId:        bson.NewObjectId(),
		usingEquips:   NewUsingEquips(),
		items:         NewItems(acc.world.Configs().maxCharItems),
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

func (c *Char) saveChar(accs *mgo.Collection) {
	ci := strconv.Itoa(c.slotIndex)
	cii := "chars." + ci
	update := bson.M{"$set": bson.M{cii: c.DumpDB()}}
	if err := accs.UpdateId(c.account.bsonId, update); err != nil {
		panic(err)
	}
}

func (c *Char) DoSaveByAccountDB() {
	c.saveChar(c.account.db.accounts)
}

func (c *Char) DoSave() {
	c.saveChar(c.db.accounts)
}

func (c *Char) Save() {
	c.job <- func() {
		c.DoSave()
	}
}

func (c *Char) DumpDB() *CharDumpDB {
	cDump := &CharDumpDB{
		Id:          c.bsonId,
		SlotIndex:   c.slotIndex,
		Name:        c.name,
		Level:       c.level,
		Str:         c.str,
		Vit:         c.vit,
		Wis:         c.wis,
		Spi:         c.spi,
		Items:       c.items.DumpDB(),
		UsingEquips: c.usingEquips.DumpDB(),
		LastScene:   nil,
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
		c.account.world.logger.Println("Char:", c.name, "logouted.")
	}
}
