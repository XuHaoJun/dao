package dao

import (
	"github.com/xuhaojun/chipmunk"
	"github.com/xuhaojun/chipmunk/vect"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"reflect"
	"strconv"
	"time"
)

var (
	CharLayer = chipmunk.Layer(8)
)

type CharSkillLevel int
type CharSkillBaseId int
type CharLearnedSkills map[CharSkillBaseId]CharSkillLevel

func (lSkills CharLearnedSkills) Client() map[string]int {
	client := make(map[string]int, len(lSkills))
	for id, level := range lSkills {
		client[strconv.Itoa(int(id))] = int(level)
	}
	return client
}

func (sid CharSkillBaseId) MaxLevel() int {
	switch sid {
	case 1:
		return 20
	}
	return -1
}

type CharClientCall interface {
	Logout()
	Move(x, y float32)
	ShutDownMove()
	TalkScene(content string)
	// equip
	EquipBySlot(slot int)
	UnequipBySlot(slot int)
	// use
	DropItem(baseId int, slotIndex int)
	PickItem(sbId int)
	UseItemBySlot(slot int)
	// NormalAttackByMid(mid int)
	// PickItemById(id int)
	// npc
	TalkNpcById(nid int)
	CancelTalkingNpc()
	ResponseTalkingNpc(optIndex int)
	//shop
	BuyItemFromOpeningShop(sellIndex int)
	SellItemToOpeningShop(baseId int, slotIndex int)
	CancelOpeningShop()
	// skill
	UseFireBall()
	UseSkillByBaseId(sid int)
	SetSkillHotKey(index int, sid int)
	SetLeftSkillHotKey(sid int)
	SetRightSkillHotKey(sid int)
	ClearNormalHotKey(index int)
	ClearSkillHotKey(index int)
	SetNormalHotKey(index int, itemBaseId int, slotIndex int)
}

type Charer interface {
	Bioer
	DumpDB() *CharDumpDB
	Login()
	Logout()
	OnReceiveClientCall(sender ClientCallPublisher, c *ClientCall)
	Save()
	SendClientCall(msg *ClientCall)
	SendClientCalls(msg []*ClientCall)
	CharClientCall() CharClientCall
	CharClient() *CharClient
	CharClientBasic() *CharClientBasic
	GetItemByBaseId(baseId int)
	Bioer() Bioer
	OpenShop(s Shoper)
	CancelOpeningShop()
	LearnSkillByBaseId(sid int)
	SendChatMessage(ch string, talkerName string, content string)
	ClientChatMessage(ch string, talkerName string, content string) *ClientCall
	SendNpcTalkBox(nt *NpcTalk)
}

type Char struct {
	*Bio
	bsonId        bson.ObjectId
	usingEquips   UsingEquips
	items         *Items // may be rename to inventory
	world         *World
	account       *Account
	slotIndex     int
	isOnline      bool
	sock          *wsConn
	lastSceneInfo *SceneInfo
	saveSceneInfo *SceneInfo
	dzeny         int
	//
	baseStr int
	baseVit int
	baseWis int
	baseSpi int
	//
	pickRadius float32
	pickRange  *chipmunk.Body
	//
	openingShop *Shop
	//
	learnedSkills CharLearnedSkills
	hotKeys       *CharHotKeys
}

type CharClient struct {
	BioClient     *BioClient        `json:"bioConfig"`
	SlotIndex     int               `json:"slotIndex"`
	LastSceneName string            `json:"lastSceneName"`
	LastX         float32           `json:"lastX"`
	LastY         float32           `json:"lastY"`
	Items         *ItemsClient      `json:"items"`
	UsingEquips   UsingEquipsClient `json:"usingEquips"`
	Dzeny         int               `json:"dzeny"`
	LearnedSkills map[string]int    `json:"learnedSkills"`
	HotKeys       *CharHotKeys      `json:"hotKeys"`
	PickRadius    float32           `json:"pickRadius"`
}

type CharClientBasic struct {
	BioClient *BioClientBasic `json:"bioConfig"`
}

type CharDumpDB struct {
	Id            bson.ObjectId     `bson:"_id"`
	SlotIndex     int               `bson:"slotIndex"`
	Name          string            `bson:"name"`
	Level         int               `bson:"level"`
	Hp            int               `bson:"hp"`
	Mp            int               `bson:"mp"`
	Str           int               `bson:"str"`
	Vit           int               `bson:"vit"`
	Wis           int               `bson:"wis"`
	Spi           int               `bson:"spi"`
	Dzeny         int               `bson:"dzeny"`
	LastScene     *SceneInfo        `bson:"lastScene"`
	SaveScene     *SceneInfo        `bson:"saveScene"`
	UsingEquips   UsingEquipsDumpDB `bson:"usingEquips"`
	Items         *ItemsDumpDB      `bson:"items"`
	BodyViewId    int               `bson:"bodyViewId"`
	BodyShape     *CircleShape      `bson:"bodyShape"`
	LearnedSkills map[string]int    `bson:"learnedSkills"`
	HotKeys       *CharHotKeys      `bson:"hotKeys"`
}

type CircleShape struct {
	Radius float32
}

type CharNormalHotKey struct {
	ItemBaseId int `json:"itemBaseId" bson:"itemBaseId"`
	SlotIndex  int `json:"slotIndex" bson:"slotIndex"`
}

type CharSkillHotKey struct {
	SkillBaseId int `json:"skillBaseId" bson:"skillBaseId"`
}

type CharHotKeys struct {
	Normal [4]*CharNormalHotKey `json:"normal" bson:"normal"`
	Skill  [2]*CharSkillHotKey  `json:"skill" bson:"skill"`
}

func NewCharHotKeys() *CharHotKeys {
	hotKeys := &CharHotKeys{
		Normal: [4]*CharNormalHotKey{},
		Skill:  [2]*CharSkillHotKey{},
	}
	for i := 0; i < 4; i++ {
		hotKeys.Normal[i] = &CharNormalHotKey{-1, -1}
	}
	for i := 0; i < 2; i++ {
		hotKeys.Skill[i] = &CharSkillHotKey{-1}
	}
	return hotKeys
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
		SaveScene:   c.saveSceneInfo,
		BodyViewId:  c.bodyViewId,
		BodyShape:   &CircleShape{32},
		HotKeys:     c.hotKeys,
	}
	cDump.LearnedSkills = map[string]int{}
	for id, level := range c.learnedSkills {
		cDump.LearnedSkills[strconv.Itoa(int(id))] = int(level)
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

func (c *Char) UseSkillByBaseId(sid int) {
	if sid <= 0 {
		return
	}
	switch sid {
	case 1:
		c.UseFireBall()
	}
}

func (c *Char) LearnSkillByBaseId(sid int) {
	// server
	if sid <= 0 {
		return
	}
	level, isLearned := c.learnedSkills[CharSkillBaseId(sid)]
	if isLearned {
		if int(level) >= int(CharSkillBaseId(sid).MaxLevel()) {
			return
		}
		c.learnedSkills[CharSkillBaseId(sid)] += 1
		learnedLevel := int(c.learnedSkills[CharSkillBaseId(sid)])
		if sid == 1 {
			c.fireBallSkill.level = learnedLevel
		}
	} else {
		c.learnedSkills[CharSkillBaseId(sid)] = 1
	}
	// client
	clientCall := &ClientCall{
		Receiver: "char",
		Method:   "handleLearnedSkills",
		Params:   []interface{}{c.learnedSkills.Client()},
	}
	c.SendClientCall(clientCall)
}

func (c *Char) TeleportBySceneName(name string, x float32, y float32) (targetScene *Scene) {
	// server
	curScene := c.scene
	targetScene = c.world.FindSceneByName(name)
	if targetScene == nil {
		return
	}
	if curScene == targetScene {
		c.SetPosition(x, y)
		clientCall := &ClientCall{
			Receiver: "char",
			Method:   "handleSetPosition",
			Params: []interface{}{map[string]float32{
				"x": x,
				"y": y,
			}},
		}
		c.SendClientCall(clientCall)
		return
	}
	c.lastSceneName = curScene.name
	c.lastId = c.id
	curScene.Remove(c.SceneObjecter())
	c.SetPosition(x, y)
	targetScene.Add(c.SceneObjecter())
	// client update
	clientCalls := make([]*ClientCall, 6)
	clientCalls[0] = &ClientCall{
		Receiver: "char",
		Method:   "handleLeaveScene",
		Params:   []interface{}{},
	}
	clientCalls[1] = &ClientCall{
		Receiver: "world",
		Method:   "handleDestroyScene",
		Params:   []interface{}{curScene.name},
	}
	clientCalls[2] = &ClientCall{
		Receiver: "world",
		Method:   "handleAddScene",
		Params:   []interface{}{targetScene.SceneClient()},
	}
	clientCalls[3] = &ClientCall{
		Receiver: "world",
		Method:   "handleRunScene",
		Params:   []interface{}{targetScene.name},
	}
	clientCalls[4] = &ClientCall{
		Receiver: "char",
		Method:   "handleJoinScene",
		Params: []interface{}{map[string]interface{}{
			"sceneName": targetScene.name,
			"id":        c.id,
		}},
	}
	clientCalls[5] = &ClientCall{
		Receiver: "char",
		Method:   "handleSetPosition",
		Params: []interface{}{map[string]float32{
			"x": x,
			"y": y,
		}},
	}
	c.SendClientCalls(clientCalls)
	return
}

func (c *Char) UpdateItemsUseSelfItemFunc() {
	for _, uItem := range c.items.useSelfItem {
		if uItem == nil {
			continue
		}
		uItem.onUse = c.world.LoadUseSelfFuncByBaseId(uItem.baseId)
	}
}

func (cDump *CharDumpDB) Load(acc *Account) *Char {
	c := NewChar(cDump.Name, acc)
	for stringId, level := range cDump.LearnedSkills {
		id, _ := strconv.Atoi(stringId)
		c.learnedSkills[CharSkillBaseId(id)] = CharSkillLevel(level)
		if id == 1 {
			c.fireBallSkill.level = level
		}
	}
	c.slotIndex = cDump.SlotIndex
	c.bsonId = cDump.Id
	c.hp = cDump.Hp
	c.mp = cDump.Mp
	c.str = cDump.Str
	c.vit = cDump.Vit
	c.wis = cDump.Wis
	c.spi = cDump.Spi
	c.dzeny = cDump.Dzeny
	c.lastSceneInfo = cDump.LastScene
	c.saveSceneInfo = cDump.SaveScene
	c.items = cDump.Items.Load()
	c.UpdateItemsUseSelfItemFunc()
	c.usingEquips = cDump.UsingEquips.Load()
	c.bodyViewId = cDump.BodyViewId
	c.body = chipmunk.NewBody(1, 1)
	circle := chipmunk.NewCircle(vect.Vector_Zero, cDump.BodyShape.Radius)
	circle.Group = BioGroup
	circle.Layer = BioLayer | CharLayer
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
	c.Bio.skillUser = c
	c.Bio.clientCallPublisher = c
	c.Bio.beKilleder = c
	c.viewAOIState.body.SetPosition(c.body.Position())
	c.CalcAttributes()
	c.hotKeys = cDump.HotKeys
	return c
}

func NewChar(name string, acc *Account) *Char {
	dConfig := acc.world.DaoConfigs()
	c := &Char{
		Bio:         NewBio(acc.world),
		bsonId:      bson.NewObjectId(),
		usingEquips: NewUsingEquips(),
		items:       NewItems(dConfig.CharConfigs.MaxCharItems),
		isOnline:    false,
		account:     acc,
		world:       acc.world,
		sock:        acc.sock,
		lastSceneInfo: &SceneInfo{
			dConfig.CharConfigs.FirstScene.Name,
			dConfig.CharConfigs.FirstScene.X,
			dConfig.CharConfigs.FirstScene.Y,
		},
		saveSceneInfo: &SceneInfo{
			dConfig.CharConfigs.FirstScene.Name,
			dConfig.CharConfigs.FirstScene.X,
			dConfig.CharConfigs.FirstScene.Y,
		},
		learnedSkills: map[CharSkillBaseId]CharSkillLevel{},
		hotKeys:       NewCharHotKeys(),
		pickRadius:    111.0,
	}
	for _, shape := range c.body.Shapes {
		shape.Layer = shape.Layer | CharLayer
	}
	c.dzeny = acc.world.DaoConfigs().CharConfigs.InitDzeny
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
	// c.pickRadius = 42.0
	// c.pickRange = chipmunk.NewBody(1, 1)
	// pickShape := chipmunk.NewCircle(vect.Vector_Zero, c.pickRadius)
	// pickShape.IsSensor = true
	// c.pickRange.IgnoreGravity = true
	// c.pickRange.AddShape(pickShape)
	c.OnKill = c.OnKillFunc()
	c.viewAOIState.OnSceneObjectEnter = c.OnSceneObjectEnterViewAOIFunc()
	c.viewAOIState.OnSceneObjectLeave = c.OnSceneObjectLeaveViewAOIFunc()
	c.body.UserData = c
	c.clientCallPublisher = c
	c.Bio.skillUser = c
	c.Bio.beKilleder = c
	c.fireBallSkill.ballLayer = MobLayer
	return c
}

func (c *Char) ClearNormalHotKey(index int) {
	c.hotKeys.Normal[index].ItemBaseId = -1
	c.hotKeys.Normal[index].SlotIndex = -1
	clientCall := &ClientCall{
		Receiver: "char",
		Method:   "handleUpdateConfig",
		Params: []interface{}{
			struct {
				CharHotKeys *CharHotKeys `json:"hotKeys"`
			}{c.hotKeys},
		},
	}
	c.SendClientCall(clientCall)
}

func (c *Char) SetNormalHotKey(index int, itemBaseId int, slotIndex int) {
	if index > 4 || itemBaseId <= 0 || slotIndex < 0 {
		return
	}
	c.hotKeys.Normal[index].ItemBaseId = itemBaseId
	c.hotKeys.Normal[index].SlotIndex = slotIndex
	clientCall := &ClientCall{
		Receiver: "char",
		Method:   "handleUpdateConfig",
		Params: []interface{}{
			struct {
				CharHotKeys *CharHotKeys `json:"hotKeys"`
			}{c.hotKeys},
		},
	}
	c.SendClientCall(clientCall)
}

func (c *Char) ClearSkillHotKey(index int) {
	if index >= 2 {
		return
	}
	c.hotKeys.Skill[index].SkillBaseId = -1
	clientCall := &ClientCall{
		Receiver: "char",
		Method:   "handleUpdateConfig",
		Params: []interface{}{
			struct {
				CharHotKeys *CharHotKeys `json:"hotKeys"`
			}{c.hotKeys},
		},
	}
	c.SendClientCall(clientCall)
}

func (c *Char) SetSkillHotKey(index int, sid int) {
	if sid <= 0 || index >= len(c.hotKeys.Skill) {
		return
	}
	c.hotKeys.Skill[index].SkillBaseId = sid
	clientCall := &ClientCall{
		Receiver: "char",
		Method:   "handleUpdateConfig",
		Params: []interface{}{
			struct {
				CharHotKeys *CharHotKeys `json:"hotKeys"`
			}{c.hotKeys},
		},
	}
	c.SendClientCall(clientCall)
}

func (c *Char) SetLeftSkillHotKey(sid int) {
	c.SetSkillHotKey(0, sid)
}

func (c *Char) SetRightSkillHotKey(sid int) {
	c.SetSkillHotKey(1, sid)
}

func (c *Char) UseFireBall() {
	pos := c.body.Position()
	clientCall := &ClientCall{
		Receiver: "char",
		Method:   "handleSetPosition",
		Params: []interface{}{map[string]float32{
			"x": float32(pos.X),
			"y": float32(pos.Y),
		}},
	}
	c.sock.SendClientCall(clientCall)
	c.Bio.UseFireBall()
	c.Bio.ShutDownMove()
	c.world.logger.Println(c.name + " use fire ball")
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

func (c *Char) GetInitItems() {
	initItems := c.world.DaoConfigs().CharConfigs.InitItems
	if initItems == nil {
		return
	}
	for _, itemPair := range initItems {
		itemBaseId := itemPair[0]
		itemCount := itemPair[1]
		if itemCount == 0 {
			c.GetItemByBaseId(itemBaseId)
		} else {
			for i := 0; i < itemCount; i++ {
				c.GetItemByBaseId(itemBaseId)
			}
		}
	}
}

func (c *Char) CharClient() *CharClient {
	bClient := c.Bio.BioClient()
	learnedSkills := map[string]int{}
	for id, level := range c.learnedSkills {
		learnedSkills[strconv.Itoa(int(id))] = int(level)
	}
	return &CharClient{
		BioClient:     bClient,
		SlotIndex:     c.slotIndex,
		LastSceneName: c.lastSceneInfo.Name,
		LastX:         c.lastSceneInfo.X,
		LastY:         c.lastSceneInfo.Y,
		Items:         c.items.ItemsClient(),
		UsingEquips:   c.usingEquips.UsingEquipsClient(),
		Dzeny:         c.dzeny,
		HotKeys:       c.hotKeys,
		PickRadius:    c.pickRadius,
		LearnedSkills: learnedSkills,
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

func (c *Char) PickItem(sbId int) {
	scene := c.scene
	if scene == nil {
		return
	}
	item := scene.FindItemerById(sbId)
	if item == nil || reflect.ValueOf(item).IsNil() {
		return
	}
	charPos := c.body.Position()
	itemPos := item.Body().Position()
	dist := vect.Dist(charPos, itemPos)
	if float32(dist) > c.pickRadius {
		return
	}
	scene.Remove(item)
	item, slotIndex := c.GetItem(item)
	if slotIndex == -1 {
		return
	}
	itemsUpdate := make(map[string]interface{}, 1)
	itemsUpdate[strconv.Itoa(slotIndex)] = item.Client()
	itemsClientUpdate := map[string]interface{}{
		item.ItemTypeByBaseId(): itemsUpdate,
	}
	clientCalls := make([]*ClientCall, 2)
	clientCalls[0] = &ClientCall{
		Receiver: "char",
		Method:   "handleUpdateItems",
		Params:   []interface{}{itemsClientUpdate},
	}
	clientCalls[1] = &ClientCall{
		Receiver: "char",
		Method:   "handleSetPosition",
		Params: []interface{}{map[string]float32{
			"x": float32(charPos.X),
			"y": float32(charPos.Y),
		}},
	}
	c.sock.SendClientCalls(clientCalls)
}

func (c *Char) ClientChatMessage(ch string, talkerName string, content string) *ClientCall {
	clientCall := &ClientCall{
		Receiver: "char",
		Method:   "handleChatMessage",
		Params: []interface{}{
			&ChatMessageClient{
				time.Now(),
				ch,
				talkerName,
				content,
			},
		},
	}
	return clientCall
}

func (c *Char) SendChatMessage(ch string, talkerName string, content string) {
	c.sock.SendClientCall(c.ClientChatMessage(ch, talkerName, content))
}

func (c *Char) SendNpcTalkBox(nt *NpcTalk) {
	var client *NpcTalkClient = nil
	if nt != nil {
		client = nt.NpcTalkClient()
	}
	clientCall := &ClientCall{
		Receiver: "char",
		Method:   "handleNpcTalkBox",
		Params:   []interface{}{client},
	}
	c.sock.SendClientCall(clientCall)
}

func (c *Char) DropItem(id int, slotIndex int) {
	if id <= 0 || slotIndex < 0 {
		return
	}
	item := c.items.RemoveItem(id, slotIndex)
	if item == nil || reflect.ValueOf(item).IsNil() {
		return
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	body := item.Body()
	pos := c.body.Position()
	rX := vect.Float(r.Intn(100))
	rY := vect.Float(r.Intn(100))
	if r.Float32() <= 0.5 {
		rY *= -1
	}
	if r.Float32() > 0.5 {
		rX *= -1
	}
	pos.X += vect.Float(rX)
	pos.Y += vect.Float(rY)
	body.SetPosition(pos)
	c.scene.Add(item.SceneObjecter())
	// client
	itemsUpdate := make(map[string]interface{}, 1)
	itemsUpdate[strconv.Itoa(slotIndex)] = nil
	itemsClientUpdate := map[string]interface{}{
		ItemTypeByBaseId(id): itemsUpdate,
	}
	clientCalls := make([]*ClientCall, 2)
	clientCalls[0] = &ClientCall{
		Receiver: "char",
		Method:   "handleUpdateItems",
		Params:   []interface{}{itemsClientUpdate},
	}
	charPos := c.body.Position()
	clientCalls[1] = &ClientCall{
		Receiver: "char",
		Method:   "handleSetPosition",
		Params: []interface{}{map[string]float32{
			"x": float32(charPos.X),
			"y": float32(charPos.Y),
		}},
	}
	c.sock.SendClientCalls(clientCalls)
}

func (c *Char) Login() {
	if c.isOnline == true ||
		c.lastSceneInfo == nil ||
		c.saveSceneInfo == nil {
		return
	}
	c.isOnline = true
	var scene *Scene
	lastScene, foundLast := c.world.scenes[c.lastSceneInfo.Name]
	if foundLast == false {
		saveScene, foundSave := c.world.scenes[c.saveSceneInfo.Name]
		if foundSave == false {
			c.saveSceneInfo = &SceneInfo{"daoCity", 0, 0}
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
		mob, ok := target.(Mober)
		if !ok {
			return
		}
		c.dzeny += mob.Level() * 100
		clientCall := &ClientCall{
			Receiver: "char",
			Method:   "handleUpdateConfig",
			Params: []interface{}{
				struct {
					Dzeny int `json:"dzeny"`
				}{c.dzeny},
			},
		}
		c.SendClientCall(clientCall)
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
		if enterSb.Scene() != c.scene {
			return
		}
		// TODO
		// display new sceneobject to client screen
		// and and mober
		switch enter := enterSb.(type) {
		case Itemer:
			clientCall := &ClientCall{
				Receiver: "scene",
				Method:   "handleAddItem",
				Params:   []interface{}{enter.ItemClient()},
			}
			c.sock.SendClientCall(clientCall)
		case Npcer:
			clientCall := &ClientCall{
				Receiver: "scene",
				Method:   "handleAddNpc",
				Params:   []interface{}{enter.NpcClientBasic()},
			}
			c.sock.SendClientCall(clientCall)
		case Charer:
			if enter != c.Charer() {
				clientCall := &ClientCall{
					Receiver: "scene",
					Method:   "handleAddChar",
					Params:   []interface{}{enter.CharClientBasic()},
				}
				c.sock.SendClientCall(clientCall)
			}
		case Mober:
			clientCall := &ClientCall{
				Receiver: "scene",
				Method:   "handleAddMob",
				Params:   []interface{}{enter.MobClientBasic()},
			}
			c.sock.SendClientCall(clientCall)
		case *FireBallState:
			// c.world.logger.Println("fire ball in view aoi!")
			clientCall := &ClientCall{
				Receiver: "scene",
				Method:   "handleAddFireBall",
				Params:   []interface{}{enter.Client()},
			}
			c.sock.SendClientCall(clientCall)
		}
	}
}

func (c *Char) OnSceneObjectLeaveViewAOIFunc() func(SceneObjecter) {
	return func(leaveSb SceneObjecter) {
		curScene := leaveSb.Scene()
		if leaveSb == c.SceneObjecter() {
			return
		}
		scene := curScene
		id := leaveSb.Id()
		if leaveSb.LastSceneName() != "" {
			foundScene := c.world.FindSceneByName(leaveSb.LastSceneName())
			if foundScene != nil {
				scene = foundScene
				id = leaveSb.LastId()
			}
		}
		if scene == nil {
			return
		}
		if c.scene != nil && c.scene.name != scene.name {
			return
		}
		// remove sceneobject to client screen
		clientCall := &ClientCall{
			Receiver: "scene",
			Method:   "handleRemoveById",
			Params:   []interface{}{id, scene.name},
		}
		c.sock.SendClientCall(clientCall)
	}
}

func (c *Char) SendClientCall(msg *ClientCall) {
	c.sock.SendClientCall(msg)
}

func (c *Char) SendClientCalls(msg []*ClientCall) {
	c.sock.SendClientCalls(msg)
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
	if c.scene == nil {
		return
	}
	//   workaround way: skip self,
	// should use another way like add timeStap.
	if cc.Method == "handleMoveStateChange" &&
		cc.Params[0] == c.id {
		return
	}
	if cc.Method == "handleChatMessage" &&
		cc.Params[0].(*ChatMessageClient).ChatType != "Local" {
		c.sock.SendClientCall(cc)
		return
	}
	sb, ok := publisher.(SceneObjecter)
	if !ok {
		return
	}
	switch sb.(type) {
	case Bioer:
		_, found := c.viewAOIState.inAreaSceneObjecters[sb]
		if found {
			c.sock.SendClientCall(cc)
			return
		}
	default:
		return
	}
}

func (c *Char) PublishClientCall(cc *ClientCall) {
	if c.scene == nil {
		return
	}
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
	btype := 0
	switch e.etype {
	case Sword:
		if c.usingEquips.RightHand() == nil {
			c.usingEquips.SetRightHand(e)
			hasEquiped = true
			btype = RightHand
		} else if c.usingEquips.LeftHand() == nil {
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
	usingEquipsClientUpdate[strconv.Itoa(btype)] = e.EquipmentClient()
	clientCalls[2] = &ClientCall{
		Receiver: "char",
		Method:   "handleUpdateUsingEquips",
		Params:   []interface{}{usingEquipsClientUpdate},
	}
	c.sock.SendClientCalls(clientCalls)
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
EachEq:
	for i, isEq := range c.items.equipment {
		if isEq == nil {
			itemsEquipSlot = i
			c.items.equipment[i] = eq
			hasUnequiped = true
			break EachEq
		}
	}
	if hasUnequiped == false {
		return
	}
	c.CalcAttributes()
	// client update
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
	c.sock.SendClientCalls(clientCalls)
}

func (c *Char) UpdateClientItems() {
}

func (c *Char) SetTalkingNpcInfo(tNpc *TalkingNpcInfo) {
	c.talkingNpcInfo = tNpc
	if tNpc == nil {
		c.SendNpcTalkBox(nil)
	}
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
	// client update
	clientCall := &ClientCall{
		Receiver: "char",
		Method:   "handleNpcTalkBox",
		Params:   []interface{}{npc.FirstNpcTalkClient()},
	}
	c.SendClientCall(clientCall)
}

// FIXME
// add npc selectoption method
// func (c *Char) ResponseTalkingNpc(optIndex int) {
// c.Bio.ResponseTalkingNpc(optIndex)
// if optIndex < 0 || c.talkingNpcInfo.target == nil {
// 	return
// }
// npc := c.talkingNpcInfo.target
// npc.SelectOption(optIndex, c.Bioer())
// talkingOpts := c.talkingNpcInfo.options
// c.talkingNpcInfo.options = append(talkingOpts, optIndex)
// }

func (c *Char) ResponseTalkingNpc(optIndex int) {
	if optIndex < 0 || c.talkingNpcInfo.target == nil {
		return
	}
	npc := c.talkingNpcInfo.target
	npc.SelectOption(optIndex, c.Bioer())
}

func (c *Char) CancelTalkingNpc() {
	c.Bio.CancelTalkingNpc()
	// client update
	clientCall := &ClientCall{
		Receiver: "char",
		Method:   "handleNpcTalkBox",
		Params:   []interface{}{nil},
	}
	c.SendClientCall(clientCall)
}

// TODO
// add error
func (c *Char) GetItem(item Itemer) (Itemer, int) {
	switch item.ItemTypeByBaseId() {
	case "equipment":
		for i, eq := range c.items.equipment {
			if eq == nil {
				c.items.equipment[i] = item.(*Equipment)
				return item, i
			}
		}
	case "useSelfItem":
		for i, us := range c.items.useSelfItem {
			if us != nil && us.baseId == item.BaseId() &&
				us.stackCount < us.maxStackCount {
				us.stackCount += 1
				return us.Itemer(), i
			} else if us == nil {
				c.items.useSelfItem[i] = item.(*UseSelfItem)
				return item, i
			}
		}
	case "etcItem":
		for i, etc := range c.items.etcItem {
			if etc != nil && etc.baseId == item.BaseId() &&
				etc.stackCount < etc.maxStackCount {
				etc.stackCount += 1
				return etc.Itemer(), i
			} else if etc == nil {
				c.items.etcItem[i] = item.(*EtcItem)
				return item, i
			}
		}
	}
	return nil, -1
}

func (c *Char) GetItemByBaseId(baseId int) {
	baseItem, err := c.world.NewItemByBaseId(baseId)
	if err != nil {
		return
	}
	item, putedSlot := c.GetItem(baseItem)
	if putedSlot == -1 {
		return
	}
	itemsUpdate := make(map[string]interface{})
	iType := item.ItemTypeByBaseId()
	itemsUpdate[strconv.Itoa(putedSlot)] = item.Client()
	// client update
	itemsClientUpdate := map[string]interface{}{
		iType: itemsUpdate,
	}
	clientCall := &ClientCall{
		Receiver: "char",
		Method:   "handleUpdateItems",
		Params:   []interface{}{itemsClientUpdate},
	}
	c.SendClientCall(clientCall)
}

func (c *Char) OpenShop(s Shoper) {
	if c.openingShop != nil {
		return
	}
	shop := s.Shop()
	c.openingShop = shop
	clientCall := &ClientCall{
		Receiver: "char",
		Method:   "handleShop",
		Params:   []interface{}{shop.ShopClient()},
	}
	c.SendClientCall(clientCall)
}

func (c *Char) UseItemBySlot(slot int) {
	if slot < 0 {
		return
	}
	uitem := c.items.useSelfItem[slot]
	if uitem == nil {
		return
	}
	useFunc := uitem.OnUseFunc()
	if useFunc != nil {
		useFunc(c.Bioer())
	} else {
		c.world.logger.Println("useFunc is nil")
	}
	uitem.stackCount -= 1
	if uitem.stackCount < 0 {
		c.items.useSelfItem[slot] = nil
	}
	// client update
	putedSlot := strconv.Itoa(slot)
	itemsUpdate := make(map[string]interface{})
	iType := uitem.ItemTypeByBaseId()
	if uitem.stackCount < 0 {
		itemsUpdate[putedSlot] = nil
	} else {
		itemsUpdate[putedSlot] = map[string]int{
			"stackCount": uitem.stackCount + 1,
		}
	}
	itemsClientUpdate := map[string]interface{}{
		iType: itemsUpdate,
	}
	clientCall := &ClientCall{
		Receiver: "char",
		Method:   "handleUpdateItems",
		Params:   []interface{}{itemsClientUpdate, true},
	}
	c.SendClientCall(clientCall)
}

func (c *Char) BuyItemFromOpeningShop(i int) {
	shop := c.openingShop
	if i < 0 || shop == nil {
		return
	}
	baseItem := shop.NewItemBySellIndex(i)
	if c.dzeny < baseItem.BuyPrice() || baseItem == nil {
		return
	}
	c.dzeny -= baseItem.BuyPrice()
	item, putedSlot := c.GetItem(baseItem)
	// client update
	itemsUpdate := make(map[string]interface{})
	iType := item.ItemTypeByBaseId()
	itemsUpdate[strconv.Itoa(putedSlot)] = item.Client()
	itemsClientUpdate := map[string]interface{}{
		iType: itemsUpdate,
	}
	clientCalls := make([]*ClientCall, 2)
	clientCalls[0] = &ClientCall{
		Receiver: "char",
		Method:   "handleUpdateItems",
		Params:   []interface{}{itemsClientUpdate},
	}
	clientCalls[1] = &ClientCall{
		Receiver: "char",
		Method:   "handleUpdateConfig",
		Params: []interface{}{
			map[string]int{"dzeny": c.dzeny},
		},
	}
	c.SendClientCalls(clientCalls)
}

func (c *Char) SellItemToOpeningShop(baseId int, slotIndex int) {
	logger := c.world.logger
	if baseId <= 0 ||
		slotIndex < 0 || slotIndex >= 30 ||
		c.openingShop == nil {
		return
	}
	var foundItem Itemer
	iType := ItemTypeByBaseId(baseId)
	switch iType {
	case "equipment":
		foundItem = c.items.equipment[slotIndex]
	case "etcItem":
		foundItem = c.items.etcItem[slotIndex]
	case "useSelfItem":
		foundItem = c.items.useSelfItem[slotIndex]
	}
	if reflect.ValueOf(foundItem).IsNil() {
		logger.Println("detect foundItem is nil")
		return
	}
	logger.Println("foundItem: ", foundItem)
	c.dzeny += foundItem.SellPrice()
	var finalItem Itemer
	switch iType {
	case "equipment":
		c.items.equipment[slotIndex] = nil
		finalItem = nil
	case "useSelfItem":
		us := c.items.useSelfItem[slotIndex]
		us.stackCount -= 1
		if us.stackCount < 0 {
			c.items.useSelfItem[slotIndex] = nil
			finalItem = nil
		} else {
			finalItem = us
		}
	case "etcItem":
		etc := c.items.etcItem[slotIndex]
		etc.stackCount -= 1
		if etc.stackCount < 0 {
			c.items.etcItem[slotIndex] = nil
			finalItem = nil
		} else {
			finalItem = etc
		}
	}
	// update client
	itemsUpdate := make(map[string]interface{})
	if finalItem != nil {
		itemsUpdate[strconv.Itoa(slotIndex)] = finalItem.Client()
	} else {
		itemsUpdate[strconv.Itoa(slotIndex)] = nil
	}
	itemsClientUpdate := map[string]interface{}{
		ItemTypeByBaseId(baseId): itemsUpdate,
	}
	clientCalls := make([]*ClientCall, 2)
	clientCalls[0] = &ClientCall{
		Receiver: "char",
		Method:   "handleUpdateItems",
		Params:   []interface{}{itemsClientUpdate},
	}
	clientCalls[1] = &ClientCall{
		Receiver: "char",
		Method:   "handleUpdateConfig",
		Params: []interface{}{
			map[string]int{"dzeny": c.dzeny},
		},
	}
	c.SendClientCalls(clientCalls)
}

func (c *Char) CancelOpeningShop() {
	if c.openingShop == nil {
		return
	}
	c.openingShop = nil
	clientCall := &ClientCall{
		Receiver: "char",
		Method:   "handleShop",
		Params:   []interface{}{nil},
	}
	c.SendClientCall(clientCall)
}

func (c *Char) TakeDamage(d *BattleDamage, attacker Bioer) {
	c.Bio.TakeDamage(d, attacker)
	clientCall := &ClientCall{
		Receiver: "char",
		Method:   "handleUpdateConfig",
		Params: []interface{}{map[string]int{
			"hp":    c.hp,
			"mp":    c.mp,
			"maxHp": c.maxHp,
			"maxSp": c.maxMp,
		}},
	}
	c.SendClientCall(clientCall)
	if !c.IsDied() {
		return
	}
	c.world.BioReborn <- c.Bioer()
	clientCalls := make([]*ClientCall, 2)
	clientCalls[0] = &ClientCall{
		Receiver: "char",
		Method:   "handleLeaveScene",
		Params:   []interface{}{},
	}
	clientCalls[1] = &ClientCall{
		Receiver: "world",
		Method:   "handleDestroyScene",
		Params:   []interface{}{c.lastSceneName},
	}
	c.SendClientCalls(clientCalls)
}

func (c *Char) Reborn() {
	scene := c.world.FindSceneByName(c.saveSceneInfo.Name)
	if scene == nil || c.scene != nil {
		return
	}
	c.hp = c.maxHp
	c.SetPosition(c.saveSceneInfo.X, c.saveSceneInfo.Y)
	scene.Add(c.SceneObjecter())
	// client
	clientCalls := make([]*ClientCall, 5)
	sceneParam := scene.SceneClient()
	clientCalls[0] = &ClientCall{
		Receiver: "world",
		Method:   "handleAddScene",
		Params:   []interface{}{sceneParam},
	}
	clientCalls[1] = &ClientCall{
		Receiver: "world",
		Method:   "handleRunScene",
		Params:   []interface{}{scene.name},
	}
	char := c
	charParam := map[string]interface{}{
		"sceneName": scene.name,
		"id":        char.id,
	}
	clientCalls[2] = &ClientCall{
		Receiver: "char",
		Method:   "handleSetPosition",
		Params: []interface{}{map[string]float32{
			"x": c.saveSceneInfo.X,
			"y": c.saveSceneInfo.Y,
		}},
	}
	clientCalls[3] = &ClientCall{
		Receiver: "char",
		Method:   "handleJoinScene",
		Params:   []interface{}{charParam},
	}
	clientCalls[4] = &ClientCall{
		Receiver: "char",
		Method:   "handleUpdateConfig",
		Params: []interface{}{map[string]int{
			"hp":    c.hp,
			"mp":    c.mp,
			"maxHp": c.maxHp,
			"maxSp": c.maxMp,
		}},
	}
	c.SendClientCalls(clientCalls)
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
