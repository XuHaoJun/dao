package dao

import (
	"strconv"

	"github.com/xuhaojun/chipmunk"
	"github.com/xuhaojun/chipmunk/vect"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type CharClientCall interface {
	Logout()
	Move(x, y float32)
	ShutDownMove()
	TalkScene(content string)
	EquipBySlot(slot int)
	UnequipBySlot(slot int)
	// NormalAttackByMid(mid int)
	// PickItemById(id int)
	TalkNpcById(nid int)
	CancelTalkingNpc()
}

type Charer interface {
	Bioer
	DumpDB() *CharDumpDB
	Login()
	Logout()
	OnReceiveClientCall(sender ClientCallPublisher, c *ClientCall)
	Save()
	SendMsg(interface{})
	CharClientCall() CharClientCall
	CharClient() *CharClient
	CharClientBasic() *CharClientBasic
	Bioer() Bioer
	ResponseTalkingNpc(optIndex int)
}

type Char struct {
	*Bio
	bsonId      bson.ObjectId
	usingEquips UsingEquips
	items       *Items // may be rename to inventory
	world       *World
	account     *Account
	slotIndex   int
	isOnline    bool
	sock        *wsConn
	lastScene   *SceneInfo
	saveScene   *SceneInfo
	dzeny       int
	//
	baseStr int
	baseVit int
	baseWis int
	baseSpi int
	//
	pickRadius float32
	pickRange  *chipmunk.Body
}

type CharClient struct {
	BioClient     *BioClient        `json:"bioConfig"`
	SlotIndex     int               `json:"slotIndex"`
	LastSceneName string            `json:"lastSceneName"`
	LastX         float32           `json:"lastX"`
	LastY         float32           `json:"lastY"`
	Items         *ItemsClient      `json:"items"`
	UsingEquips   UsingEquipsClient `json:"usingEquips"`
}

type CharClientBasic struct {
	BioClient *BioClientBasic `json:"bioConfig"`
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
	c.body = chipmunk.NewBody(1, 1)
	circle := chipmunk.NewCircle(vect.Vector_Zero, cDump.BodyShape.Radius)
	circle.Group = BioGroup
	circle.SetFriction(0)
	circle.SetElasticity(0)
	c.body.AddShape(circle)
	c.body.SetPosition(vect.Vect{
		X: vect.Float(cDump.LastScene.X),
		Y: vect.Float(cDump.LastScene.Y)})
	c.body.IgnoreGravity = true
	c.body.SetVelocity(0, 0)
	c.body.SetMoment(chipmunk.Inf)
	c.body.UserData = c
	c.viewAOIState.body.SetPosition(c.body.Position())
	c.CalcAttributes()
	return c
}

func NewChar(name string, acc *Account) *Char {
	c := &Char{
		Bio:         NewBio(acc.world),
		bsonId:      bson.NewObjectId(),
		usingEquips: NewUsingEquips(),
		items:       NewItems(acc.world.Configs().maxCharItems),
		isOnline:    false,
		account:     acc,
		world:       acc.world,
		sock:        acc.sock,
		lastScene:   &SceneInfo{"daoCity", 0, 0},
		saveScene:   &SceneInfo{"daoCity", 0, 0},
	}
	c.name = name
	c.level = 1
	c.baseStr = 1
	c.baseVit = 1
	c.baseWis = 1
	c.baseSpi = 1
	c.str = 1
	c.vit = 1
	c.wis = 1
	c.spi = 1
	c.CalcAttributes()
	c.hp = c.maxHp
	c.mp = c.maxMp
	freeEq, err := c.world.NewEquipmentByBaseId(1)
	if err == nil {
		c.items.equipment[0] = freeEq
	}
	c.pickRadius = 42.0
	c.pickRange = chipmunk.NewBody(1, 1)
	pickShape := chipmunk.NewCircle(vect.Vector_Zero, c.pickRadius)
	pickShape.IsSensor = true
	c.pickRange.IgnoreGravity = true
	c.pickRange.AddShape(pickShape)
	c.OnKill = c.OnKillFunc()
	c.viewAOIState.OnSceneObjectEnter = c.OnSceneObjectEnterViewAOIFunc()
	c.viewAOIState.OnSceneObjectLeave = c.OnSceneObjectLeaveViewAOIFunc()
	c.body.UserData = c
	c.clientCallPublisher = c
	return c
}

func (c *Char) CalcAttributes() {
	for _, eq := range c.usingEquips {
		if eq == nil {
			continue
		}
		c.str = c.baseStr + eq.bonusInfo.str
		c.wis = c.baseWis + eq.bonusInfo.wis
		c.spi = c.baseSpi + eq.bonusInfo.spi
		c.vit = c.baseVit + eq.bonusInfo.vit
	}
	c.Bio.CalcAttributes()
	for _, eq := range c.usingEquips {
		if eq == nil {
			continue
		}
		c.maxHp += eq.bonusInfo.maxHp
		c.maxMp += eq.bonusInfo.maxHp
		c.atk += eq.bonusInfo.atk
		c.matk += eq.bonusInfo.matk
		c.def += eq.bonusInfo.def
		c.mdef += eq.bonusInfo.mdef
	}
}

func (c *Char) CharClient() *CharClient {
	bClient := c.Bio.BioClient()
	return &CharClient{
		BioClient:     bClient,
		SlotIndex:     c.slotIndex,
		LastSceneName: c.lastScene.Name,
		LastX:         c.lastScene.X,
		LastY:         c.lastScene.Y,
		Items:         c.items.ItemsClient(),
		UsingEquips:   c.usingEquips.UsingEquipsClient(),
	}
}

func (c *Char) CharClientBasic() *CharClientBasic {
	bClient := c.Bio.BioClientBasic()
	return &CharClientBasic{
		BioClient: bClient,
	}
}

func (c *Char) CharClientCall() CharClientCall {
	return c
}

func (c *Char) saveChar(accs *mgo.Collection) {
	ci := strconv.Itoa(c.slotIndex)
	cii := "chars." + ci
	update := bson.M{"$set": bson.M{cii: c.DumpDB()}}
	if err := accs.UpdateId(c.account.bsonId, update); err != nil {
		panic(err)
	}
}

func (c *Char) Save() {
	c.saveChar(c.account.world.db.accounts)
}

func (c *Char) Login() {
	if c.isOnline == true ||
		c.lastScene == nil ||
		c.saveScene == nil {
		return
	}
	c.isOnline = true
	var scene *Scene
	lastScene, foundLast := c.world.scenes[c.lastScene.Name]
	if foundLast == false {
		saveScene, foundSave := c.world.scenes[c.saveScene.Name]
		if foundSave == false {
			c.saveScene = &SceneInfo{"daoCity", 0, 0}
			scene = c.world.scenes["daoCity"]
		} else {
			scene = saveScene
		}
	} else {
		scene = lastScene
	}
	scene.Add(c.SceneObjecter())
	logger := c.account.world.logger
	logger.Println("Char:", c.name, "logined.")
}

func (c *Char) OnKillFunc() func(target Bioer) {
	return func(target Bioer) {
		// for quest check or add zeny when kill
		mob, ok := target.(*Mob)
		if !ok {
			return
		}
		c.dzeny += mob.Level() * 10
	}
}

func (c *Char) NormalAttackByMid(mid int) {
	if mid <= 0 || c.scene == nil {
		return
	}
	mob := c.scene.FindMobById(mid)
	if mob == nil {
		return
	}
	// c.NormalAttack(mob.Bioer())
}

func (c *Char) Logout() {
	c.account.Logout()
}

func (c *Char) OnSceneObjectEnterViewAOIFunc() func(SceneObjecter) {
	return func(enterSb SceneObjecter) {
		// TODO
		// display new sceneobject to client screen
		// and and mober
		switch enter := enterSb.(type) {
		case Npcer:
			clientCall := &ClientCall{
				Receiver: "scene",
				Method:   "handleAddNpc",
				Params:   []interface{}{enter.NpcClientBasic()},
			}
			c.sock.SendMsg(clientCall)
		case Charer:
			if enter != c.Charer() {
				clientCall := &ClientCall{
					Receiver: "scene",
					Method:   "handleAddChar",
					Params:   []interface{}{enter.CharClientBasic()},
				}
				c.sock.SendMsg(clientCall)
			}
		}
	}
}

func (c *Char) OnSceneObjectLeaveViewAOIFunc() func(SceneObjecter) {
	return func(leaveSb SceneObjecter) {
		// remove sceneobject to client screen
		clientCall := &ClientCall{
			Receiver: "scene",
			Method:   "handleRemoveById",
			Params:   []interface{}{leaveSb.Id()},
		}
		c.sock.SendMsg(clientCall)
	}
}

func (c *Char) SendMsg(msg interface{}) {
	c.sock.SendMsg(msg)
}

func (c *Char) Bioer() Bioer {
	return c
}

func (c *Char) Charer() Charer {
	return c
}

func (c *Char) SceneObjecter() SceneObjecter {
	return c
}

// TODO
// will add timeStamp for better.
// it dispatch from scene
func (c *Char) OnReceiveClientCall(publisher ClientCallPublisher, cc *ClientCall) {
	//   workaround way: skip self,
	// should use another way like add timeStap.
	if cc.Method == "handleMoveStateChange" &&
		cc.Params[0] == c.id {
		return
	}
	if cc.Method == "handleChatMessage" &&
		cc.Params[0].(*ChatMessageClient).ChatType != "Local" {
		c.sock.SendMsg(cc)
		return
	}
	sb, ok := publisher.(SceneObjecter)
	if !ok {
		return
	}
	switch realSb := sb.(type) {
	case *Char:
		_, found := c.viewAOIState.inAreaSceneObjecters[realSb]
		if found {
			c.sock.SendMsg(cc)
			return
		}
	default:
		return
	}
}

func (c *Char) PublishClientCall(cc *ClientCall) {
	c.scene.DispatchClientCall(c, cc)
}

func (c *Char) EquipBySlot(slot int) {
	if slot < 0 {
		return
	}
	e := c.items.equipment[slot]
	if e == nil {
		return
	}
	c.items.equipment[slot] = nil
	hasEquiped := false
	etype := 0
	switch e.etype {
	case Sword:
		if c.usingEquips.LeftHand() == nil {
			c.usingEquips.SetLeftHand(e)
			hasEquiped = true
			etype = LeftHand
		} else if c.usingEquips.RightHand() == nil {
		}
	}
	if hasEquiped == false {
		return
	}
	c.CalcAttributes()
	// update client
	clientCalls := make([]*ClientCall, 3)
	clientCalls[0] = &ClientCall{
		Receiver: "char",
		Method:   "handleUpdateConfig",
		Params:   []interface{}{c.BioClientAttributes()},
	}
	itemsEqUpdate := make(map[string]interface{})
	itemsEqUpdate[strconv.Itoa(slot)] = nil
	itemsClientUpdate := struct {
		Equipment map[string]interface{} `json:"equipment"`
	}{
		itemsEqUpdate,
	}
	clientCalls[1] = &ClientCall{
		Receiver: "char",
		Method:   "handleUpdateItems",
		Params:   []interface{}{itemsClientUpdate},
	}
	usingEquipsClientUpdate := make(map[string]interface{})
	usingEquipsClientUpdate[strconv.Itoa(etype)] = e.EquipmentClient()
	clientCalls[2] = &ClientCall{
		Receiver: "char",
		Method:   "handleUpdateUsingEquips",
		Params:   []interface{}{usingEquipsClientUpdate},
	}
	c.sock.SendMsg(clientCalls)
}

func (c *Char) UnequipBySlot(slot int) {
	if slot < 0 || slot > 11 || c.usingEquips[slot] == nil {
		return
	}
	// FIXME
	eq := c.usingEquips[slot]
	c.usingEquips[slot] = nil
	hasUnequiped := false
	itemsEquipSlot := 0
	for i, isEq := range c.items.equipment {
		if isEq == nil {
			itemsEquipSlot = i
			c.items.equipment[i] = eq
			hasUnequiped = true
			break
		}
	}
	if hasUnequiped == false {
		return
	}
	c.CalcAttributes()
	// update client
	clientCalls := make([]*ClientCall, 3)
	clientCalls[0] = &ClientCall{
		Receiver: "char",
		Method:   "handleUpdateConfig",
		Params:   []interface{}{c.BioClientAttributes()},
	}
	itemsEqUpdate := make(map[string]interface{})
	itemsEqUpdate[strconv.Itoa(itemsEquipSlot)] = eq.EquipmentClient()
	itemsClientUpdate := struct {
		Equipment map[string]interface{} `json:"equipment"`
	}{
		itemsEqUpdate,
	}
	clientCalls[1] = &ClientCall{
		Receiver: "char",
		Method:   "handleUpdateItems",
		Params:   []interface{}{itemsClientUpdate},
	}
	usingEquipsClientUpdate := make(map[string]interface{})
	usingEquipsClientUpdate[strconv.Itoa(slot)] = nil
	clientCalls[2] = &ClientCall{
		Receiver: "char",
		Method:   "handleUpdateUsingEquips",
		Params:   []interface{}{usingEquipsClientUpdate},
	}
	c.sock.SendMsg(clientCalls)
}

func (c *Char) TalkNpcById(nid int) {
	if nid <= 0 || c.scene == nil {
		return
	}
	npc := c.scene.FindNpcerById(nid)
	if npc == nil {
		return
	}
	isFirsted := npc.FirstBeTalked(c.Bioer())
	if isFirsted == false {
		return
	}
	clientCall := &ClientCall{
		Receiver: "char",
		Method:   "handleNpcTalkBox",
		Params:   []interface{}{npc.NpcTalkClient()},
	}
	c.SendMsg(clientCall)
}

func (c *Char) ResponseTalkingNpc(optIndex int) {
	if optIndex < 0 || c.talkingNpcInfo.target == nil {
		return
	}
	npc := c.talkingNpcInfo.target
	npc.SelectOption(optIndex, c.Bioer())
	talkingOpts := c.talkingNpcInfo.options
	c.talkingNpcInfo.options = append(talkingOpts, optIndex)
}

func (c *Char) CancelTalkingNpc() {
	c.Bio.CancelTalkingNpc()
	clientCall := &ClientCall{
		Receiver: "char",
		Method:   "handleNpcTalkBox",
		Params:   []interface{}{nil},
	}
	c.SendMsg(clientCall)
}

// func (c *Char) DoAddItem(itemer Itemer) {
// 	switch item := itemer.(type) {
// 	case *EtcItem:
// 		for i, eitem := range c.items.etcItem {
// 			if eitem == nil {
// 				c.items.etcItem[i] = item
// 				break
// 			}
// 		}
// 	case *Equipment:
// 		for i, eitem := range c.items.equipment {
// 			if eitem == nil {
// 				c.items.equipment[i] = item
// 				break
// 			}
// 		}
// 	case *UseSelfItem:
// 		for i, uitem := range c.items.useSelfItem {
// 			if uitem == nil {
// 				c.items.useSelfItem[i] = item
// 				break
// 			}
// 		}
// 	}
// }

// func (c *Char) AddItem(itemer Itemer) {
// 	c.DoJob(func() {
// 		c.DoAddItem(itemer)
// 	})
// }

// func (c *Char) DoPickItem(item Itemer) {
// 	item.Lock()
// 	defer item.Unlock()
// 	if item.GetScene() == nil {
// 		return
// 	}
// 	c.DoAddItem(item)
// 	item.GetScene().DeleteItem(item)
// 	item.DoSetScene(nil)
// }

// // TODO
// // func add pick item check func

// func (c *Char) PickItem(item Itemer) {
// 	c.DoJob(func() {
// 		c.DoPickItem(item)
// 	})
// }

// func (c *Char) PickItemById(id int) {
// 	c.DoJob(func() {
// 		if id < 0 {
// 			return
// 		}
// 		item := c.scene.FindItemId(id)
// 		if item == nil {
// 			return
// 		}
// 		c.DoPickItem(item)
// 		c.world.logger.Println("Char:", c.name, "pick up", item.Name())
// 	})
// }

// func (c *Char) MoveByXY(x float64, y float64) {
// 	c.Move(vect.Vect{X: vect.Float(x), Y: vect.Float(y)})
// }
