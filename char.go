package dao

import (
	"strconv"

	"github.com/xuhaojun/chipmunk"
	"github.com/xuhaojun/chipmunk/vect"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type CharClientCall interface {
	Logout()
	NormalAttackByMid(mid int)
	PickItemById(id int)
}

type Char struct {
	*BattleBioBase
	bsonId      bson.ObjectId
	usingEquips UsingEquips
	items       *Items // may be rename to inventory
	world       *World
	account     *Account
	slotIndex   int
	db          *DaoDB
	isOnline    bool
	sock        *wsConn
	lastScene   *SceneInfo
	saveScene   *SceneInfo
	dzeny       int
	//
	pickRadius float32
	pickRange  *chipmunk.Body
}

type CharClient struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	SlotIndex     int    `json:"slotIndex"`
	LastSceneName string `json:"lastSceneName"`
	BodyViewId    int    `json:"bodyViewid"`
	Level         int    `json:"level"`
	IsDied        bool   `json:"isDied"`
	// main attribue
	Str int `json:"str"`
	Vit int `json:"vit"`
	Wis int `json:"wis"`
	Spi int `json:"spi"`
	// sub attribue
	Def   int `json:"def"`
	Mdef  int `json:"mdef"`
	Atk   int `json:"atk"`
	Matk  int `json:"matk"`
	MaxHp int `json:"maxHp"`
	Hp    int `json:"hp"`
	MaxMp int `json:"maxMp"`
	Mp    int `json:"mp"`
	// TODO
	// add items and body info
}

type CharDumpDB struct {
	Id          bson.ObjectId     `bson:"_id"`
	SlotIndex   int               `bson:"slotIndex"`
	Name        string            `bson:"name"`
	Level       int               `bson:"level"`
	Hp          int               `bson:"hp"`
	Mp          int               `bson:"mp"`
	Str         int               `bson:"str"`
	Vit         int               `bson:"vit"`
	Wis         int               `bson:"wis"`
	Spi         int               `bson:"spi"`
	Dzeny       int               `bson:"dzeny"`
	LastScene   *SceneInfo        `bson:"lastScene"`
	SaveScene   *SceneInfo        `bson:"saveScene"`
	UsingEquips UsingEquipsDumpDB `bson:"usingEquips"`
	Items       *ItemsDumpDB      `bson:"items"`
	BodyViewId  int               `bson:"bodyViewId"`
	BodyShape   *CircleShape      `bson:"bodyShape"`
}

type CircleShape struct {
	Radius float32
}

func (c *Char) DumpDB() *CharDumpDB {
	cDump := &CharDumpDB{
		Id:          c.bsonId,
		SlotIndex:   c.slotIndex,
		Name:        c.name,
		Level:       c.level,
		Hp:          c.hp,
		Mp:          c.mp,
		Str:         c.str,
		Vit:         c.vit,
		Wis:         c.wis,
		Spi:         c.spi,
		Dzeny:       c.dzeny,
		Items:       c.items.DumpDB(),
		UsingEquips: c.usingEquips.DumpDB(),
		LastScene:   nil,
		SaveScene:   c.saveScene,
		BodyViewId:  c.bodyViewId,
		BodyShape:   &CircleShape{32},
	}
	if c.scene != nil {
		cDump.LastScene = &SceneInfo{
			c.scene.name,
			float32(c.body.Position().X),
			float32(c.body.Position().Y),
		}
	} else {
		cDump.LastScene = &SceneInfo{"daoCity", 0.0, 0.0}
	}
	return cDump
}

func (cDump *CharDumpDB) Load(acc *Account) *Char {
	c := NewChar(cDump.Name, acc)
	c.slotIndex = cDump.SlotIndex
	c.bsonId = cDump.Id
	c.hp = cDump.Hp
	c.mp = cDump.Mp
	c.str = cDump.Str
	c.vit = cDump.Vit
	c.wis = cDump.Wis
	c.spi = cDump.Spi
	c.dzeny = cDump.Dzeny
	c.lastScene = cDump.LastScene
	c.saveScene = cDump.SaveScene
	c.items = cDump.Items.Load()
	c.usingEquips = cDump.UsingEquips.Load()
	c.bodyViewId = cDump.BodyViewId
	body := chipmunk.NewBody(1, 1)
	body.IgnoreGravity = true
	body.SetVelocity(0, 0)
	body.SetMoment(chipmunk.Inf)
	c.body = body
	c.body.SetPosition(vect.Vect{
		X: vect.Float(cDump.LastScene.X),
		Y: vect.Float(cDump.LastScene.Y)})
	circle := chipmunk.NewCircle(vect.Vector_Zero, cDump.BodyShape.Radius)
	circle.SetFriction(0)
	circle.SetElasticity(0)
	c.body.AddShape(circle)
	c.DoCalcAttributes()
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
		lastScene:     &SceneInfo{"daoCity", 0, 0},
		saveScene:     &SceneInfo{"daoCity", 0, 0},
	}
	c.name = name
	c.level = 1
	c.str = 1
	c.vit = 1
	c.wis = 1
	c.spi = 1
	c.pickRadius = 42.0
	c.pickRange = chipmunk.NewBody(1, 1)
	pickShape := chipmunk.NewCircle(vect.Vector_Zero, c.pickRadius)
	pickShape.IsSensor = true
	c.pickRange.IgnoreGravity = true
	c.pickRange.AddShape(pickShape)
	// replace default onkill func
	c.BattleBioBase.OnKill = c.OnKillFunc()
	return c
}

func (c *Char) CharClient() *CharClient {
	cc := &CharClient{
		Id:            c.id,
		Name:          c.name,
		Level:         c.level,
		SlotIndex:     c.slotIndex,
		LastSceneName: c.lastScene.Name,
		BodyViewId:    c.bodyViewId,
		IsDied:        c.isDied,
		//
		Str: c.str,
		Vit: c.vit,
		Wis: c.wis,
		Spi: c.spi,
		//
		Def:   c.def,
		Mdef:  c.mdef,
		Atk:   c.atk,
		Matk:  c.matk,
		MaxHp: c.maxHp,
		Hp:    c.hp,
		MaxMp: c.maxMp,
		Mp:    c.mp,
	}
	return cc
}

func (c *Char) CharClientCall() CharClientCall {
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
			if c.scene != nil {
				c.scene.DeleteBio(c)
				c.scene = nil
				c.id = 0
			}
			c.DoLogout()
			c.quit <- struct{}{}
			return
		}
	}
}

func (c *Char) DoCalcAttributes() {
	// TODO
	// add attributes from usingEquips
	c.BattleBioBase.DoCalcAttributes()
}

func (c *Char) CalcAttributes() {
	c.DoJob(func() {
		c.DoCalcAttributes()
	})
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
	c.DoJob(func() {
		c.DoSave()
	})
}

func (c *Char) Login() {
	go c.Run()
	c.DoJob(func() {
		if c.isOnline == true ||
			c.lastScene == nil ||
			c.saveScene == nil {
			return
		}
		c.isOnline = true
		var scene *Scene
		lastScene := c.world.FindSceneByName(c.lastScene.Name)
		if lastScene == nil {
			saveScene := c.world.FindSceneByName(c.saveScene.Name)
			if saveScene == nil {
				c.saveScene = &SceneInfo{"daoCity", 0, 0}
				scene = c.world.FindSceneByName(c.saveScene.Name)
			} else {
				scene = saveScene
			}
		} else {
			scene = lastScene
		}
		scene.AddBio(c)
		logger := c.account.world.logger
		logger.Println("Char:", c.name, "logined.")
	})
}

func (c *Char) OnKillFunc() func(target BattleBioer) {
	return func(target BattleBioer) {
		// for quest check or add zeny when kill
		mob, ok := target.(*Mob)
		if !ok {
			return
		}
		c.dzeny += mob.Level() * 10
	}
}

func (c *Char) NormalAttackByMid(mid int) {
	c.DoJob(func() {
		if mid <= 0 || c.scene == nil {
			return
		}
		mob := c.scene.FindMobById(mid)
		if mob == nil {
			return
		}
		c.NormalAttack(mob.BattleBioer())
	})
}

func (c *Char) DoLogout() {
	if c.isOnline == false {
		return
	}
	c.isOnline = false
	c.DoSave()
	c.account.world.logger.Println("Char:", c.name, "logouted.")
}

func (c *Char) Logout() {
	c.account.Logout()
}

func (c *Char) OnReceiveClientCall(publisher ClientCallPublisher, cc *ClientCall) {
	c.DoJob(func() {
		// if cc.Method == "Talk" &&
		// 	len(cc.Params) == 2 &&
		// 	cc.Params[0].(string) == "local" {
		// 	c.sock.SendMsg(cc)
		// 	return
		// }
		// TODO
		// add itemPublisher in the future
		// bioPublisher, ok := publisher.(Bioer)
		// if !ok {
		// 	return
		// }
		// _, found := c.viewAOI.bioers[bioPublisher]
		// if found {
		// 	c.sock.SendMsg(cc)
		// }
	})
}

func (c *Char) DoHasEquipInUsingEquips(e *Equipment) bool {
	for i := 0; i < len(c.usingEquips); i++ {
		if e == c.usingEquips[i] {
			return true
		}
	}
	return false
}

func (c *Char) DoHasEquipInItems(e *Equipment) bool {
	for i := 0; i < len(c.items.equipment); i++ {
		if e == c.items.equipment[i] {
			return true
		}
	}
	return false
}

func (c *Char) EquipBySlot(slot int) {
	c.DoJob(func() {
		if slot < 0 {
			return
		}
		e := c.items.equipment[slot]
		if e == nil {
			return
		}
		c.DoEquip(e)
	})
}

func (c *Char) DoEquip(e *Equipment) {
	c.usingEquips[e.etype] = e
	// TODO
	// inc equip's bonus to char
	c.DoCalcAttributes()
}

func (c *Char) Equip(e *Equipment) {
	c.DoJob(func() {
		found := c.DoHasEquipInItems(e)
		if found == false {
			return
		}
		c.DoEquip(e)
	})
}

func (c *Char) DoUnequip(e *Equipment) {
	c.usingEquips[e.etype] = nil
	// TODO
	// dec equip's bonus to char
	c.DoCalcAttributes()
}

func (c *Char) Unequip(e *Equipment) {
	c.DoJob(func() {
		found := c.DoHasEquipInUsingEquips(e)
		if found == false {
			return
		}
		c.DoUnequip(e)
	})
}

func (c *Char) UnequipBySlot(slot int) {
	c.DoJob(func() {
		if slot < 0 {
			return
		}
		e := c.usingEquips[slot]
		if e == nil {
			return
		}
		c.DoUnequip(e)
	})
}

func (c *Char) DoAddItem(itemer Itemer) {
	switch item := itemer.(type) {
	case *EtcItem:
		for i, eitem := range c.items.etcItem {
			if eitem == nil {
				c.items.etcItem[i] = item
				break
			}
		}
	case *Equipment:
		for i, eitem := range c.items.equipment {
			if eitem == nil {
				c.items.equipment[i] = item
				break
			}
		}
	case *UseSelfItem:
		for i, uitem := range c.items.useSelfItem {
			if uitem == nil {
				c.items.useSelfItem[i] = item
				break
			}
		}
	}
}

func (c *Char) AddItem(itemer Itemer) {
	c.DoJob(func() {
		c.DoAddItem(itemer)
	})
}

func (c *Char) DoPickItem(item Itemer) {
	item.Lock()
	defer item.Unlock()
	if item.GetScene() == nil {
		return
	}
	c.DoAddItem(item)
	item.GetScene().DeleteItem(item)
	item.DoSetScene(nil)
}

// TODO
// func add pick item check func

func (c *Char) PickItem(item Itemer) {
	c.DoJob(func() {
		c.DoPickItem(item)
	})
}

func (c *Char) PickItemById(id int) {
	c.DoJob(func() {
		if id < 0 {
			return
		}
		item := c.scene.FindItemId(id)
		if item == nil {
			return
		}
		c.DoPickItem(item)
		c.world.logger.Println("Char:", c.name, "pick up", item.Name())
	})
}
