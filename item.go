package dao

import (
	"github.com/xuhaojun/chipmunk"
	"github.com/xuhaojun/chipmunk/vect"
)

type Itemer interface {
	SceneObjecter
	Name() string
	SetName(string)
	AgeisName() string
	SetAgeisName(string)
	IconViewId() int
	SetIconViewId(int)
	DumpDB() *ItemDumpDB
	Owner() Bioer
	SetOwner(Bioer)
}

type Item struct {
	id         int
	baseId     int
	name       string
	ageisName  string
	iconViewId int
	scene      *Scene
	owner      Bioer
	body       *chipmunk.Body
	bodyViewId int
}

type ItemDumpDB struct {
	Name       string `bson:"name"`
	AgeisName  string `bson:"ageisName"`
	IconViewId int    `bson:"iconViewid"`
	BaseId     int    `bson:"baseId"`
}

func (i *Item) DumpDB() *ItemDumpDB {
	return &ItemDumpDB{
		Name:       i.name,
		AgeisName:  i.ageisName,
		IconViewId: i.iconViewId,
		BaseId:     i.baseId,
	}
}

func (idump *ItemDumpDB) Load() *Item {
	item := NewItem()
	item.name = idump.Name
	item.ageisName = idump.AgeisName
	item.iconViewId = idump.IconViewId
	item.baseId = idump.BaseId
	return item
}

type ItemClient struct {
	Id         int           `json:"id"`
	Name       string        `json:"name"`
	AgeisName  string        `json:"ageisName"`
	IconViewId int           `json:"iconViewId"`
	CpBody     *CpBodyClient `json:"cpBody"`
	BodyViewId int           `json:"bodyViewId"`
}

func (i *Item) ItemClient() *ItemClient {
	return &ItemClient{
		Id:         i.id,
		Name:       i.name,
		AgeisName:  i.ageisName,
		IconViewId: i.iconViewId,
		CpBody:     ToCpBodyClient(i.body),
		BodyViewId: i.bodyViewId,
	}
}

func NewItem() *Item {
	circle := chipmunk.NewCircle(vect.Vector_Zero, 12.0)
	circle.Group = BioGroup
	circle.SetFriction(0)
	circle.SetElasticity(0)
	body := chipmunk.NewBody(1, 1)
	body.IgnoreGravity = true
	body.SetMoment(chipmunk.Inf)
	body.SetVelocity(0, 0)
	body.SetPosition(vect.Vector_Zero)
	body.AddShape(circle)
	return &Item{
		body:       body,
		iconViewId: 1,
		bodyViewId: 3000,
	}
}

func (i *Item) Id() int {
	return i.id
}

func (i *Item) SetId(id int) {
	i.id = id
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

func (i *Item) Scene() *Scene {
	return i.scene
}

func (i *Item) SetScene(s *Scene) {
	i.scene = s
}

func (i *Item) Owner() Bioer {
	return i.owner
}

func (i *Item) SetOwner(b Bioer) {
	i.owner = b
}

func (i *Item) Body() *chipmunk.Body {
	return i.body
}

func (i *Item) AfterUpdate(delta float32) {
}

func (i *Item) BeforeUpdate(delta float32) {
}

func (i *Item) OnBeAddedToScene(s *Scene) {
}

func (i *Item) OnBeRemovedToScene(s *Scene) {
}

type EquipmentClient struct {
	ItemClient  *ItemClient               `json:"itemConfig"`
	BonusInfo   *EquipmentBonusInfoClient `json:"bonusInfo"`
	EquipViewId int                       `json:"equipViewId"`
	EquipLimit  *EquipLimitClient         `json:"equipLimit"`
}

func (e *Equipment) EquipmentClient() *EquipmentClient {
	return &EquipmentClient{
		ItemClient:  e.ItemClient(),
		BonusInfo:   e.bonusInfo.EquipmentBonusInfoClient(),
		EquipViewId: e.equipViewId,
		EquipLimit:  e.equipLimit.EquipLimitClient(),
	}
}

type Equipment struct {
	*Item
	bonusInfo   *EquipmentBonusInfo
	etype       int
	equipViewId int
	equipLimit  *EquipLimit
}

func NewEquipment() *Equipment {
	return &Equipment{
		Item:       NewItem(),
		bonusInfo:  &EquipmentBonusInfo{},
		equipLimit: &EquipLimit{},
	}
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
		Etype:       e.etype,
		EquipViewId: e.equipViewId,
		EquipLimit:  e.equipLimit.DumpDB(),
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
		bonusInfo:   e.BonusInfo.Load(),
		etype:       e.Etype,
		equipViewId: e.EquipViewId,
		equipLimit:  e.EquipLimit.Load(),
	}
}

type EquipmentDumpDB struct {
	Item *ItemDumpDB `bson:"item"`
	//
	Etype       int `bson:"etype"`
	EquipViewId int `bson:"equipViewId"`
	//
	BonusInfo  *EquipmentBonusInfoDumpDB `bson:"bonusInfo"`
	EquipLimit *EquipLimitDumpDB         `bson:"equipLimit"`
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
	Helm = iota
	Pauldrons
	Armor
	Shield
	Golves
	Belt
	Boot
	Ring
	Amulet
	Pant
	// Weapons
	Sword
	Stick
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

func (ue UsingEquips) LeftHand() *Equipment {
	return ue[LeftHand]
}

func (ue UsingEquips) RightHand() *Equipment {
	return ue[RightHand]
}

func (ue UsingEquips) Head() *Equipment {
	return ue[Head]
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

func (ue UsingEquips) Hands() *Equipment {
	return ue[Hands]
}

func (ue UsingEquips) Waist() *Equipment {
	return ue[Waist]
}

func (ue UsingEquips) Legs() *Equipment {
	return ue[Legs]
}

func (ue UsingEquips) LeftFinger() *Equipment {
	return ue[LeftFinger]
}

func (ue UsingEquips) RightFinger() *Equipment {
	return ue[RightFinger]
}

func (ue UsingEquips) Neck() *Equipment {
	return ue[Neck]
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
		StackCount:    e.stackCount,
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

type UseSelfItem struct {
	*Item
	stackCount    int
	maxStackCount int
}

type UseSelfItemDumpDB struct {
	Item          *ItemDumpDB `bson:"item"`
	StackCount    int         `bson:"stackCount"`
	MaxStackCount int         `bson:"maxStackcount"`
}

type UseSelfItemClient struct {
	Item          *ItemClient `json:"itemConfig"`
	StackCount    int         `json:"stackCount"`
	MaxStackCount int         `json:"maxStackcount"`
}

func (u *UseSelfItem) UseSelfItemClient() *UseSelfItemClient {
	return &UseSelfItemClient{
		Item:          u.Item.ItemClient(),
		StackCount:    u.stackCount,
		MaxStackCount: u.maxStackCount,
	}
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
