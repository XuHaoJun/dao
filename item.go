package dao

import (
	"errors"
	"github.com/xuhaojun/chipmunk"
	"github.com/xuhaojun/chipmunk/vect"
	"gopkg.in/mgo.v2/bson"
	"reflect"
)

var (
	ItemLayer = chipmunk.Layer(2)
)

type Itemer interface {
	SceneObjecter
	SceneObjecter() SceneObjecter
	Name() string
	SetName(string)
	AgeisName() string
	SetAgeisName(string)
	IconViewId() int
	SetIconViewId(int)
	// DumpDB() *ItemDumpDB
	Owner() Bioer
	SetOwner(Bioer)
	BaseId() int
	SellPrice() int
	SetSellPrice(price int)
	BuyPrice() int
	ItemTypeByBaseId() string
	ItemClient() *ItemClient
	Client() interface{}
}

type Item struct {
	*SceneObject
	baseId     int
	name       string
	ageisName  string
	iconViewId int
	owner      Bioer
	bodyViewId int
	buyPrice   int
	sellPrice  int
}

type ItemDumpDB struct {
	Name       string `bson:"name"`
	AgeisName  string `bson:"ageisName"`
	IconViewId int    `bson:"iconViewId"`
	BaseId     int    `bson:"baseId"`
	BuyPrice   int    `bson:"buyPrice"`
	SellPrice  int    `bson:"sellPrice"`
}

func (i *Item) DumpDB() *ItemDumpDB {
	return &ItemDumpDB{
		Name:       i.name,
		AgeisName:  i.ageisName,
		IconViewId: i.iconViewId,
		BaseId:     i.baseId,
		BuyPrice:   i.buyPrice,
		SellPrice:  i.sellPrice,
	}
}

func (i *Item) Client() interface{} {
	return i.ItemClient()
}

func (idump *ItemDumpDB) Load() *Item {
	item := NewItem()
	item.name = idump.Name
	item.ageisName = idump.AgeisName
	item.iconViewId = idump.IconViewId
	item.baseId = idump.BaseId
	item.buyPrice = idump.BuyPrice
	item.sellPrice = idump.SellPrice
	return item
}

type ItemClient struct {
	Id         int           `json:"id"`
	BaseId     int           `json:"baseId"`
	Name       string        `json:"name"`
	AgeisName  string        `json:"ageisName"`
	IconViewId int           `json:"iconViewId"`
	CpBody     *CpBodyClient `json:"cpBody"`
	BodyViewId int           `json:"bodyViewId"`
	BuyPrice   int           `json:"buyPrice"`
	SellPrice  int           `json:"sellPrice"`
}

func (i *Item) SceneObjecter() SceneObjecter {
	return i
}

func (i *Item) ItemClient() *ItemClient {
	return &ItemClient{
		Id:         i.id,
		BaseId:     i.baseId,
		Name:       i.name,
		AgeisName:  i.ageisName,
		IconViewId: i.iconViewId,
		CpBody:     ToCpBodyClient(i.body),
		BodyViewId: i.bodyViewId,
		BuyPrice:   i.buyPrice,
		SellPrice:  i.sellPrice,
	}
}

func NewItem() *Item {
	item := &Item{
		iconViewId: 1,
		bodyViewId: 3000,
	}
	circle := chipmunk.NewCircle(vect.Vector_Zero, 12.0)
	circle.Group = BioGroup
	circle.Layer = ItemLayer
	circle.SetFriction(0)
	circle.SetElasticity(0)
	body := chipmunk.NewBody(1, 1)
	body.IgnoreGravity = true
	body.SetMoment(chipmunk.Inf)
	body.SetVelocity(0, 0)
	body.SetPosition(vect.Vector_Zero)
	body.AddShape(circle)
	body.UserData = item
	item.SceneObject = &SceneObject{
		body: body,
	}
	return item
}

func (i *Item) BodyViewId() int {
	return i.bodyViewId
}

func (i *Item) Name() string {
	return i.name
}

func (i *Item) AgeisName() string {
	return i.ageisName
}

func (i *Item) SetName(name string) {
	i.name = name
}

func (i *Item) SetAgeisName(name string) {
	i.ageisName = name
}

func (i *Item) IconViewId() int {
	return i.iconViewId
}

func (i *Item) SetIconViewId(vid int) {
	i.iconViewId = vid
}

func (i *Item) Owner() Bioer {
	return i.owner
}

func (i *Item) SetOwner(b Bioer) {
	i.owner = b
}

func (i *Item) BuyPrice() int {
	return i.buyPrice
}

func (i *Item) SetSellPrice(price int) {
	i.sellPrice = price
}

func (i *Item) SellPrice() int {
	return i.sellPrice
}

func (i *Item) AfterUpdate(delta float32) {
}

func (i *Item) BeforeUpdate(delta float32) {
}

func (i *Item) OnBeAddedToScene(s *Scene) {
}

func (i *Item) OnBeRemovedToScene(s *Scene) {
}

func ItemTypeByBaseId(baseId int) string {
	var iType string
	if baseId >= 1 && baseId <= 5000 {
		iType = "equipment"
	} else if baseId >= 5001 && baseId <= 10000 {
		iType = "useSelfItem"
	} else {
		iType = "etcItem"
	}
	return iType
}

func (i *Item) ItemTypeByBaseId() string {
	return ItemTypeByBaseId(i.baseId)
}

func (i *Item) BaseId() int {
	return i.baseId
}

type EquipmentClient struct {
	ItemClient  *ItemClient               `json:"itemConfig"`
	Level       int                       `json:"level"`
	BonusInfo   *EquipmentBonusInfoClient `json:"bonusInfo"`
	EquipViewId int                       `json:"equipViewId"`
	EquipLimit  *EquipLimitClient         `json:"equipLimit"`
}

func (e *Equipment) EquipmentClient() *EquipmentClient {
	return &EquipmentClient{
		ItemClient:  e.ItemClient(),
		Level:       e.level,
		BonusInfo:   e.bonusInfo.EquipmentBonusInfoClient(),
		EquipViewId: e.equipViewId,
		EquipLimit:  e.equipLimit.EquipLimitClient(),
	}
}

type Equipment struct {
	*Item
	level       int
	bonusInfo   *EquipmentBonusInfo
	etype       int
	equipViewId int
	equipLimit  *EquipLimit
}

func (e *Equipment) Itemer() Itemer {
	return e
}

func (e *Equipment) SceneObjecter() SceneObjecter {
	return e
}

func (e *Equipment) Client() interface{} {
	return e.EquipmentClient()
}

func NewEquipment() *Equipment {
	eq := &Equipment{
		level:      1,
		bonusInfo:  &EquipmentBonusInfo{},
		equipLimit: &EquipLimit{},
	}
	item := NewItem()
	item.body.UserData = eq
	eq.Item = item
	return eq
}

func (e *Equipment) Etype() int {
	return e.etype
}

func (e *Equipment) EquipViewId() int {
	return e.equipViewId
}

func (e *Equipment) BonusInfo() *EquipmentBonusInfo {
	return e.bonusInfo
}

func (e *Equipment) EquipLimit() *EquipLimit {
	return e.equipLimit
}

func (e *Equipment) DumpDB() *EquipmentDumpDB {
	return &EquipmentDumpDB{
		Item:        e.Item.DumpDB(),
		BonusInfo:   e.bonusInfo.DumpDB(),
		Level:       e.level,
		Etype:       e.etype,
		EquipViewId: e.equipViewId,
		EquipLimit:  e.equipLimit.DumpDB(),
	}
}

func (e *Equipment) DB() *EquipmentDB {
	return &EquipmentDB{
		Item:        e.Item.DumpDB(),
		BonusInfo:   e.bonusInfo.DB(),
		Level:       e.level,
		Etype:       e.etype,
		EquipViewId: e.equipViewId,
		EquipLimit:  e.equipLimit.DumpDB(),
	}
}

func (eDB *EquipmentDB) DumpDB() *EquipmentDumpDB {
	return &EquipmentDumpDB{
		Item:        eDB.Item,
		Level:       eDB.Level,
		BonusInfo:   eDB.BonusInfo.DumpDB(),
		Etype:       eDB.Etype,
		EquipViewId: eDB.EquipViewId,
		EquipLimit:  eDB.EquipLimit,
	}
}

func (eDB *EquipmentBonusInfoDB) randIt(v []int) int {
	length := len(v)
	if length == 1 {
		return v[0]
	} else if length == 2 {
		return RandIntnRange(v[0], v[1])
	}
	return 0
}

func (bDB *EquipmentBonusInfoDB) DumpDB() *EquipmentBonusInfoDumpDB {
	return &EquipmentBonusInfoDumpDB{
		MaxHp: bDB.randIt(bDB.MaxHp),
		MaxMp: bDB.randIt(bDB.MaxMp),
		Str:   bDB.randIt(bDB.Str),
		Vit:   bDB.randIt(bDB.Vit),
		Wis:   bDB.randIt(bDB.Wis),
		Spi:   bDB.randIt(bDB.Spi),
		Atk:   bDB.randIt(bDB.Atk),
		Matk:  bDB.randIt(bDB.Matk),
		Def:   bDB.randIt(bDB.Def),
		Mdef:  bDB.randIt(bDB.Mdef),
	}
}

func (e *EquipmentDumpDB) Load() *Equipment {
	if e.Item == nil {
		e.Item = NewItem().DumpDB()
	}
	if e.BonusInfo == nil {
		e.BonusInfo = NewEquipmentBonusInfo().DumpDB()
	}
	if e.EquipLimit == nil {
		e.EquipLimit = NewEquipLimit().DumpDB()
	}
	return &Equipment{
		Item:        e.Item.Load(),
		level:       e.Level,
		bonusInfo:   e.BonusInfo.Load(),
		etype:       e.Etype,
		equipViewId: e.EquipViewId,
		equipLimit:  e.EquipLimit.Load(),
	}
}

type EquipmentDumpDB struct {
	Item *ItemDumpDB `bson:"item"`
	//
	Level       int `bson:"level"`
	Etype       int `bson:"etype"`
	EquipViewId int `bson:"equipViewId"`
	//
	BonusInfo  *EquipmentBonusInfoDumpDB `bson:"bonusInfo"`
	EquipLimit *EquipLimitDumpDB         `bson:"equipLimit"`
}

type EquipmentDB struct {
	Item  *ItemDumpDB `bson:"item"`
	Level int         `bson:"level"`
	//
	Etype       int `bson:"etype"`
	EquipViewId int `bson:"equipViewId"`
	//
	BonusInfo  *EquipmentBonusInfoDB `bson:"bonusInfo"`
	EquipLimit *EquipLimitDumpDB     `bson:"equipLimit"`
}

type EquipmentBonusInfoDB struct {
	MaxHp []int `bson:"maxHp"`
	MaxMp []int `bson:"maxMp"`
	Str   []int `bson:"str"`
	Vit   []int `bson:"vit"`
	Wis   []int `bson:"wis"`
	Spi   []int `bson:"spi"`
	Atk   []int `bson:"atk"`
	Matk  []int `bson:"matk"`
	Def   []int `bson:"def"`
	Mdef  []int `bson:"mdef"`
}

func NewEquipmentBonusInfo() *EquipmentBonusInfo {
	return &EquipmentBonusInfo{}
}

func NewEquipLimit() *EquipLimit {
	return &EquipLimit{}
}

type EquipmentBonusInfo struct {
	maxHp int
	maxMp int
	str   int
	vit   int
	wis   int
	spi   int
	atk   int
	matk  int
	def   int
	mdef  int
}

type BaseEquipmentDB struct {
	Id        int                       `bson:"id"`
	Name      string                    `bson:"name"`
	Etype     int                       `bson:"etype"`
	AgeisName string                    `bson:"ageisName"`
	BonusInfo *EquipmentBonusInfoDumpDB `bson:"bonusInfo"`
}

type EquipmentBonusInfoDumpDB struct {
	MaxHp int `bson:"maxHp" json:"maxHp"`
	MaxMp int `bson:"maxMp" json:"maxMp"`
	Str   int `bson:"str" json:"str"`
	Vit   int `bson:"vit" json:"vit"`
	Wis   int `bson:"wis" json:"wis"`
	Spi   int `bson:"spi" json:"spi"`
	Atk   int `bson:"atk" json:"atk"`
	Matk  int `bson:"matk" json:"matk"`
	Def   int `bson:"def" json:"def"`
	Mdef  int `bson:"mdef" json:"mdef"`
}

type EquipmentBonusInfoClient struct {
	MaxHp int `json:"maxHp"`
	MaxMp int `json:"maxMp"`
	Str   int `json:"str"`
	Vit   int `json:"vit"`
	Wis   int `json:"wis"`
	Spi   int `json:"spi"`
	Atk   int `json:"atk"`
	Matk  int `json:"matk"`
	Def   int `json:"def"`
	Mdef  int `json:"mdef"`
}

func (b *EquipmentBonusInfo) EquipmentBonusInfoClient() *EquipmentBonusInfoClient {
	return &EquipmentBonusInfoClient{
		MaxHp: b.maxHp,
		MaxMp: b.maxMp,
		Str:   b.str,
		Vit:   b.vit,
		Wis:   b.wis,
		Spi:   b.spi,
		Atk:   b.atk,
		Matk:  b.matk,
		Def:   b.def,
		Mdef:  b.mdef,
	}
}

func (b *EquipmentBonusInfo) DumpDB() *EquipmentBonusInfoDumpDB {
	return &EquipmentBonusInfoDumpDB{
		MaxHp: b.maxHp,
		MaxMp: b.maxMp,
		Str:   b.str,
		Vit:   b.vit,
		Wis:   b.wis,
		Spi:   b.spi,
		Atk:   b.atk,
		Matk:  b.matk,
		Def:   b.def,
		Mdef:  b.mdef,
	}
}

func (b *EquipmentBonusInfo) DB() *EquipmentBonusInfoDB {
	return &EquipmentBonusInfoDB{}
}

func (b *EquipmentBonusInfoDumpDB) Load() *EquipmentBonusInfo {
	return &EquipmentBonusInfo{
		maxHp: b.MaxHp,
		maxMp: b.MaxMp,
		str:   b.Str,
		vit:   b.Vit,
		wis:   b.Wis,
		spi:   b.Spi,
		atk:   b.Atk,
		matk:  b.Matk,
		def:   b.Def,
		mdef:  b.Mdef,
	}
}

type EquipLimitDumpDB struct {
	Level int
	Str   int
	Vit   int
	Wis   int
	Spi   int
}

type EquipLimitClient EquipLimitDumpDB

type EquipLimit struct {
	level int
	str   int
	vit   int
	wis   int
	spi   int
}

func (e *EquipLimit) EquipLimitClient() *EquipLimitClient {
	return &EquipLimitClient{
		Level: e.level,
		Str:   e.str,
		Vit:   e.vit,
		Wis:   e.wis,
		Spi:   e.spi,
	}
}

func (e *EquipLimit) DumpDB() *EquipLimitDumpDB {
	return &EquipLimitDumpDB{
		Level: e.level,
		Str:   e.str,
		Vit:   e.vit,
		Wis:   e.wis,
		Spi:   e.spi,
	}
}

func (e *EquipLimitDumpDB) Load() *EquipLimit {
	return &EquipLimit{
		level: e.Level,
		str:   e.Str,
		vit:   e.Vit,
		wis:   e.Wis,
		spi:   e.Spi,
	}
}

// equip type
const (
	// Armors
	Helm      = 0
	Pauldrons = 1
	Armor     = 2
	Shield    = 3
	Golves    = 4
	Belt      = 5
	HandGuard = 6
	Ring      = 7
	Amulet    = 8
	Pants     = 9
	// Weapons
	Sword = 10
	Stick = 11
)

// equip part on body type
const (
	LeftHand = iota
	RightHand
	Head
	Shoulders
	Torso
	Wrists
	Hands
	Waist
	Legs
	LeftFinger
	RightFinger
	Neck
	MaxUsingEquip
)

type UsingEquips []*Equipment

func (ue UsingEquips) SetLeftHand(e *Equipment) {
	ue[LeftHand] = e
}

func (ue UsingEquips) SetHead(e *Equipment) {
	ue[Head] = e
}

func (ue UsingEquips) SetTorso(e *Equipment) {
	ue[Torso] = e
}

func (ue UsingEquips) SetRightHand(e *Equipment) {
	ue[RightHand] = e
}

func (ue UsingEquips) LeftHand() *Equipment {
	return ue[LeftHand]
}

func (ue UsingEquips) RightHand() *Equipment {
	return ue[RightHand]
}

func (ue UsingEquips) Head() *Equipment {
	return ue[Head]
}

func (ue UsingEquips) SetShoulders(e *Equipment) {
	ue[Shoulders] = e
}

func (ue UsingEquips) Shoulders() *Equipment {
	return ue[Shoulders]
}

func (ue UsingEquips) Torso() *Equipment {
	return ue[Torso]
}

func (ue UsingEquips) Wrists() *Equipment {
	return ue[Wrists]
}

func (ue UsingEquips) SetWrists(e *Equipment) {
	ue[Wrists] = e
}

func (ue UsingEquips) SetHands(e *Equipment) {
	ue[Hands] = e
}

func (ue UsingEquips) Hands() *Equipment {
	return ue[Hands]
}

func (ue UsingEquips) Waist() *Equipment {
	return ue[Waist]
}

func (ue UsingEquips) SetWaist(e *Equipment) {
	ue[Waist] = e
}

func (ue UsingEquips) Legs() *Equipment {
	return ue[Legs]
}

func (ue UsingEquips) SetLegs(e *Equipment) {
	ue[Legs] = e
}

func (ue UsingEquips) LeftFinger() *Equipment {
	return ue[LeftFinger]
}

func (ue UsingEquips) RightFinger() *Equipment {
	return ue[RightFinger]
}

func (ue UsingEquips) SetLeftFinger(e *Equipment) {
	ue[LeftFinger] = e
}

func (ue UsingEquips) SetRightFinger(e *Equipment) {
	ue[RightFinger] = e
}

func (ue UsingEquips) Neck() *Equipment {
	return ue[Neck]
}

func (ue UsingEquips) SetNeck(e *Equipment) {
	ue[Neck] = e
}

type UsingEquipsDumpDB []*EquipmentDumpDB

type UsingEquipsClient []*EquipmentClient

func NewUsingEquips() UsingEquips {
	es := make([]*Equipment, MaxUsingEquip)
	return UsingEquips(es)
}

func (es UsingEquips) UsingEquipsClient() UsingEquipsClient {
	esClient := make([]*EquipmentClient, MaxUsingEquip)
	for i, e := range es {
		if e != nil {
			esClient[i] = e.EquipmentClient()
		} else {
			esClient[i] = nil
		}
	}
	return UsingEquipsClient(esClient)
}

func (es UsingEquips) DumpDB() UsingEquipsDumpDB {
	esDump := make([]*EquipmentDumpDB, len(es))
	for i, e := range es {
		if e != nil {
			esDump[i] = e.DumpDB()
		}
	}
	return UsingEquipsDumpDB(esDump)
}

func (esDump UsingEquipsDumpDB) Load() UsingEquips {
	es := NewUsingEquips()
	for i, e := range esDump {
		if e != nil {
			es[i] = e.Load()
		}
	}
	return es
}

type EtcItem struct {
	*Item
	stackCount    int
	maxStackCount int
}

func (e *EtcItem) Itemer() Itemer {
	return e
}

func (e *EtcItem) SceneObjecter() SceneObjecter {
	return e
}

func (e *EtcItem) StackCount() int {
	return e.stackCount
}

func (e *EtcItem) Client() interface{} {
	return e.EtcItemClient()
}

func NewEtcItem() *EtcItem {
	etc := &EtcItem{
		maxStackCount: 1024,
	}
	item := NewItem()
	item.body.UserData = etc
	etc.Item = item
	return etc
}

type EtcItemDumpDB struct {
	Item          *ItemDumpDB `bson:"item"`
	StackCount    int         `bson:"stackCount"`
	MaxStackCount int         `bson:"maxStackCount"`
}

type EtcItemClient struct {
	Item          *ItemClient `json:"itemConfig"`
	StackCount    int         `json:"stackCount"`
	MaxStackCount int         `json:"maxStackCount"`
}

func (e *EtcItem) EtcItemClient() *EtcItemClient {
	return &EtcItemClient{
		Item:          e.Item.ItemClient(),
		StackCount:    e.stackCount + 1,
		MaxStackCount: e.maxStackCount,
	}
}

func (e *EtcItem) DumpDB() *EtcItemDumpDB {
	return &EtcItemDumpDB{
		Item:          e.Item.DumpDB(),
		StackCount:    e.stackCount,
		MaxStackCount: e.maxStackCount,
	}
}

func (e *EtcItemDumpDB) Load() *EtcItem {
	return &EtcItem{
		Item:          e.Item.Load(),
		stackCount:    e.StackCount,
		maxStackCount: e.MaxStackCount,
	}
}

type UseSelfItemer interface {
	Itemer
	OnUseFunc() func(b Bioer)
}

func (u *UseSelfItem) OnUseFunc() func(Bioer) {
	return u.onUse
}

type UseSelfItem struct {
	*Item
	onUse         func(b Bioer)
	stackCount    int
	maxStackCount int
}

func (u *UseSelfItem) StackCount() int {
	return u.stackCount
}

func (u *UseSelfItem) Client() interface{} {
	return u.UseSelfItemClient()
}

func (u *UseSelfItem) SceneObjecter() SceneObjecter {
	return u
}

func NewUseSelfItem() *UseSelfItem {
	use := &UseSelfItem{}
	item := NewItem()
	item.body.UserData = use
	use.Item = item
	return use
}

type UseSelfItemCall struct {
	Receiver string        `json:"receiver"`
	Method   string        `json:"method"`
	Params   []interface{} `json:"params"`
}

func (uCall *UseSelfItemCall) Eval(item Itemer, bio Bioer) ([]reflect.Value, error) {
	f := uCall.FindFunc(item, bio)
	if f.IsNil() {
		return nil, errors.New("eval function not found")
	}
	in, err := uCall.CastParams(f, item, bio)
	if err != nil {
		return nil, err
	}
	return f.Call(in), nil
}

func (uCall *UseSelfItemCall) FindFunc(item Itemer, bio Bioer) (f reflect.Value) {
	var receiver interface{}
	switch uCall.Receiver {
	case "Bio":
		receiver = bio
	case "Util":
		receiver = bio.World().util
	case "World":
		receiver = bio.World()
	case "Scene":
		receiver = bio.Scene()
	case "Item":
		receiver = item
	case "Char":
		char, isChar := bio.(Charer)
		if isChar {
			receiver = char
		}
	default:
		return
	}
	f = reflect.ValueOf(receiver).MethodByName(uCall.Method)
	return
}

func (uCall *UseSelfItemCall) CastParams(f reflect.Value, item Itemer, bio Bioer) ([]reflect.Value, error) {
	if f.IsValid() == false {
		return nil, errors.New("wrong function")
	}
	numIn := f.Type().NumIn()
	if len(uCall.Params) != numIn {
		return nil, errors.New("not match params length")
	}
	in := make([]reflect.Value, numIn)
	var ftype reflect.Type
	for i, param := range uCall.Params {
		ftype = f.Type().In(i)
		switch ftype.Kind() {
		case reflect.Int:
			switch param.(type) {
			case int:
				in[i] = reflect.ValueOf(param)
			case float64:
				in[i] = reflect.ValueOf(int(param.(float64)))
			case (bson.M):
				useSelfItemCallMap := param.(bson.M)
				nextCall := &UseSelfItemCall{
					Receiver: useSelfItemCallMap["receiver"].(string),
					Method:   useSelfItemCallMap["method"].(string),
					Params:   useSelfItemCallMap["params"].([]interface{}),
				}
				nextF := nextCall.FindFunc(item, bio)
				nextIn, err := nextCall.CastParams(nextF, item, bio)
				if err != nil {
					return nil, err
				}
				nextValue := nextF.Call(nextIn)
				in[i] = nextValue[0]
			default:
				return nil, errors.New("nothing not match params type int")
			}
		case reflect.String:
			switch param.(type) {
			case (bson.M):
				useSelfItemCallMap := param.(bson.M)
				nextCall := &UseSelfItemCall{
					Receiver: useSelfItemCallMap["receiver"].(string),
					Method:   useSelfItemCallMap["method"].(string),
					Params:   useSelfItemCallMap["params"].([]interface{}),
				}
				nextF := nextCall.FindFunc(item, bio)
				nextIn, err := nextCall.CastParams(nextF, item, bio)
				if err != nil {
					return nil, err
				}
				nextValue := nextF.Call(nextIn)
				in[i] = nextValue[0]
			case string:
				in[i] = reflect.ValueOf(param)
			default:
				return nil, errors.New("not match params type string")
			}
		case reflect.Float32:
			switch param.(type) {
			case int:
				in[i] = reflect.ValueOf(float32(param.(int)))
			case float32:
				in[i] = reflect.ValueOf(param)
			case float64:
				in[i] = reflect.ValueOf(float32(param.(float64)))
			default:
				return nil, errors.New("not match params type float32")
			}
		case reflect.Float64:
			switch param.(type) {
			case float32:
				in[i] = reflect.ValueOf(param.(float64))
			case float64:
				in[i] = reflect.ValueOf(param)
			default:
				return nil, errors.New("not match params type float64")
			}
		case reflect.Slice:
			switch ftype.String() {
			case "[]int":
				switch param.(type) {
				case []int:
					in[i] = reflect.ValueOf(param)
				default:
					return nil, errors.New("not match params type []int")
				}
			case "[]float64":
				switch param.(type) {
				case []float64:
					in[i] = reflect.ValueOf(param)
				default:
					return nil, errors.New("not match params type []float64")
				}
			case "[]float32":
				switch v := param.(type) {
				case []float64:
					f32s := make([]float32, len(v))
					for i, f64 := range v {
						f32s[i] = float32(f64)
					}
					in[i] = reflect.ValueOf(f32s)
				default:
					return nil, errors.New("not match params type []float32")
				}
			case "[]string":
				switch param.(type) {
				case []string:
					in[i] = reflect.ValueOf(param)
				default:
					return nil, errors.New("not match params type []string")
				}
			}
		}
	}
	return in, nil
}

type UseSelfItemDumpDB struct {
	Item          *ItemDumpDB `bson:"item"`
	StackCount    int         `bson:"stackCount"`
	MaxStackCount int         `bson:"maxStackcount"`
	//
	UseSelfFuncArrays []*UseSelfItemCall `bson:"useSelfFuncs,omitempty"`
}

type UseSelfItemClient struct {
	Item          *ItemClient `json:"itemConfig"`
	StackCount    int         `json:"stackCount"`
	MaxStackCount int         `json:"maxStackcount"`
}

func (u *UseSelfItem) UseSelfItemClient() *UseSelfItemClient {
	return &UseSelfItemClient{
		Item:          u.Item.ItemClient(),
		StackCount:    u.stackCount + 1,
		MaxStackCount: u.maxStackCount,
	}
}

func (u *UseSelfItem) Itemer() Itemer {
	return u
}

func (u *UseSelfItem) DumpDB() *UseSelfItemDumpDB {
	return &UseSelfItemDumpDB{
		Item:          u.Item.DumpDB(),
		StackCount:    u.stackCount,
		MaxStackCount: u.maxStackCount,
	}
}

func (u *UseSelfItemDumpDB) Load() *UseSelfItem {
	return &UseSelfItem{
		Item:          u.Item.Load(),
		stackCount:    u.StackCount,
		maxStackCount: u.MaxStackCount,
	}
}

type Items struct {
	equipment   []*Equipment
	etcItem     []*EtcItem
	useSelfItem []*UseSelfItem
}

func (is *Items) RemoveItem(baseId int, slotIndex int) Itemer {
	var item Itemer = nil
	switch ItemTypeByBaseId(baseId) {
	case "equipment":
		item = is.equipment[slotIndex]
		is.equipment[slotIndex] = nil
	case "etcItem":
		item = is.etcItem[slotIndex]
		is.etcItem[slotIndex] = nil
	case "useSelfItem":
		item = is.useSelfItem[slotIndex]
		is.useSelfItem[slotIndex] = nil
	}
	return item
}

func (is *Items) FindItem(baseId int, slotIndex int) Itemer {
	switch ItemTypeByBaseId(baseId) {
	case "equipment":
		return is.equipment[slotIndex]
	case "etcItem":
		return is.etcItem[slotIndex]
	case "useSelfItem":
		return is.useSelfItem[slotIndex]
	}
	return nil
}

type ItemsDumpDB struct {
	Equipment   []*EquipmentDumpDB   `bson:"equipment"`
	EtcItem     []*EtcItemDumpDB     `bson:"etcItem"`
	UseSelfItem []*UseSelfItemDumpDB `bson:"useSelfItem"`
}

type ItemsClient struct {
	Equipment   []*EquipmentClient   `json:"equipment"`
	EtcItem     []*EtcItemClient     `json:"etcItem"`
	UseSelfItem []*UseSelfItemClient `json:"useSelfItem"`
}

func (i *Items) ItemsClient() *ItemsClient {
	esClient := make([]*EquipmentClient, len(i.equipment))
	for i, e := range i.equipment {
		if e != nil {
			esClient[i] = e.EquipmentClient()
		}
	}
	eiClient := make([]*EtcItemClient, len(i.etcItem))
	for i, e := range i.etcItem {
		if e != nil {
			eiClient[i] = e.EtcItemClient()
		}
	}
	usClient := make([]*UseSelfItemClient, len(i.useSelfItem))
	for i, e := range i.useSelfItem {
		if e != nil {
			usClient[i] = e.UseSelfItemClient()
		}
	}
	return &ItemsClient{
		Equipment:   esClient,
		EtcItem:     eiClient,
		UseSelfItem: usClient,
	}
}

func (i *Items) DumpDB() *ItemsDumpDB {
	esDump := make([]*EquipmentDumpDB, len(i.equipment))
	for i, e := range i.equipment {
		if e != nil {
			esDump[i] = e.DumpDB()
		}
	}
	eiDump := make([]*EtcItemDumpDB, len(i.etcItem))
	for i, e := range i.etcItem {
		if e != nil {
			eiDump[i] = e.DumpDB()
		}
	}
	usDump := make([]*UseSelfItemDumpDB, len(i.useSelfItem))
	for i, e := range i.useSelfItem {
		if e != nil {
			usDump[i] = e.DumpDB()
		}
	}
	return &ItemsDumpDB{
		Equipment:   esDump,
		EtcItem:     eiDump,
		UseSelfItem: usDump,
	}
}

func (isDump *ItemsDumpDB) Load() *Items {
	items := NewItems(len(isDump.Equipment))
	for i, eqDump := range isDump.Equipment {
		if eqDump != nil {
			items.equipment[i] = eqDump.Load()
		}
	}
	for i, eiDump := range isDump.EtcItem {
		if eiDump != nil {
			items.etcItem[i] = eiDump.Load()
		}
	}
	for i, usDump := range isDump.UseSelfItem {
		if usDump != nil {
			items.useSelfItem[i] = usDump.Load()
		}
	}
	return items
}

func NewItems(maxItems int) *Items {
	is := &Items{
		equipment:   make([]*Equipment, maxItems),
		etcItem:     make([]*EtcItem, maxItems),
		useSelfItem: make([]*UseSelfItem, maxItems),
	}
	return is
}
