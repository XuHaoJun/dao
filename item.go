package dao

import (
	"github.com/xuhaojun/chipmunk/vect"
	"labix.org/v2/mgo/bson"
	"sync"
)

type Itemer interface {
	Name() string
	AgeisName() string
	IconViewId() int
	DoSetScene(s *Scene)
	SetScene(s *Scene)
	Scene() *Scene
	GetScene() *Scene
	SetPos(vect.Vect)
	Pos() vect.Vect
	DoJob(func())
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

// TODO
// item may be have body?

type Item struct {
	bsonId     bson.ObjectId
	name       string
	ageisName  string
	iconViewId int
	mutex      *sync.RWMutex
	scene      *Scene
	pos        vect.Vect
}

func (i *Item) DoJob(f func()) {
	i.mutex.Lock()
	f()
	i.mutex.Unlock()
}

func (i *Item) Name() string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.name
}

func (i *Item) AgeisName() string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.ageisName
}

func (i *Item) IconViewId() int {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.iconViewId
}

func (i *Item) SetScene(s *Scene) {
	i.mutex.Lock()
	i.scene = s
	i.mutex.Unlock()
}

func (i *Item) SetPos(pos vect.Vect) {
	i.mutex.Lock()
	i.pos = pos
	i.mutex.Unlock()
}

func (i *Item) Pos() vect.Vect {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.pos
}

func (i *Item) DoSetScene(s *Scene) {
	i.scene = s
}

func (i *Item) GetScene() *Scene {
	return i.scene
}

func (i *Item) Scene() *Scene {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.scene
}

func (i *Item) Lock() {
	i.mutex.Lock()
}

func (i *Item) Unlock() {
	i.mutex.Unlock()
}

func (i *Item) RLock() {
	i.mutex.RLock()
}

func (i *Item) RUnlock() {
	i.mutex.RUnlock()
}

type Equipment struct {
	*Item
	bonus       *BonusInfo
	etype       int
	equipViewId int
	equipLimit  *EquipLimit
}

func (e *Equipment) DumpDB() *EquipmentDumpDB {
	return &EquipmentDumpDB{
		Id:          e.bsonId,
		Name:        e.name,
		AgeisName:   e.ageisName,
		IconViewId:  e.iconViewId,
		Bonus:       e.bonus.DumpDB(),
		Etype:       e.etype,
		EquipViewId: e.equipViewId,
		EquipLimit:  e.equipLimit.DumpDB(),
	}
}

func (e *EquipmentDumpDB) Load() *Equipment {
	return &Equipment{
		Item: &Item{
			bsonId:     e.Id,
			name:       e.Name,
			ageisName:  e.AgeisName,
			iconViewId: e.IconViewId,
			scene:      nil,
			mutex:      &sync.RWMutex{},
			pos:        vect.Vector_Zero,
		},
		bonus:       e.Bonus.Load(),
		etype:       e.Etype,
		equipViewId: e.EquipViewId,
		equipLimit:  e.EquipLimit.Load(),
	}
}

type EquipmentDumpDB struct {
	Id          bson.ObjectId     `bson:"_id"`
	Name        string            `bson:"name"`
	AgeisName   string            `bson:"ageisName"`
	IconViewId  int               `bson:"iconViewId"`
	Bonus       *BonusInfoDumpDB  `bson:"bonus"`
	Etype       int               `bson:"etype"`
	EquipViewId int               `bson:"equipViewId"`
	EquipLimit  *EquipLimitDumpDB `bson:"equipLimit"`
}

type BonusInfo struct {
	hp    int
	maxHp int
	mp    int
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

type BonusInfoDumpDB struct {
	Hp    int
	MaxHp int
	Mp    int
	MaxMp int
	Str   int
	Vit   int
	Wis   int
	Spi   int
	Atk   int
	Matk  int
	Def   int
	Mdef  int
}

func (b *BonusInfo) DumpDB() *BonusInfoDumpDB {
	return &BonusInfoDumpDB{
		Hp:    b.hp,
		MaxHp: b.maxHp,
		Mp:    b.mp,
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

func (b *BonusInfoDumpDB) Load() *BonusInfo {
	return &BonusInfo{
		hp:    b.Hp,
		maxHp: b.MaxHp,
		mp:    b.Mp,
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

type EquipLimit struct {
	level int
	str   int
	vit   int
	wis   int
	spi   int
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
)

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

type UsingEquipsDumpDB []*EquipmentDumpDB

func NewUsingEquips() UsingEquips {
	es := make([]*Equipment, MaxUsingEquip)
	return UsingEquips(es)
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
}

type EtcItemDumpDB struct {
	Id         bson.ObjectId `bson:"_id"`
	Name       string        `bson:"name"`
	AgeisName  string        `bson:"ageisName"`
	IconViewId int           `bson:"iconViewId"`
}

func (e *EtcItem) DumpDB() *EtcItemDumpDB {
	return &EtcItemDumpDB{
		Id:         e.bsonId,
		Name:       e.name,
		AgeisName:  e.ageisName,
		IconViewId: e.iconViewId,
	}
}

func (e *EtcItemDumpDB) Load() *EtcItem {
	return &EtcItem{
		Item: &Item{
			bsonId:     e.Id,
			name:       e.Name,
			ageisName:  e.AgeisName,
			iconViewId: e.IconViewId,
		},
	}
}

type UseSelfItem struct {
	*Item
}

type UseSelfItemDumpDB struct {
	Id         bson.ObjectId `bson:"_id"`
	Name       string        `bson:"name"`
	AgeisName  string        `bson:"ageisName"`
	IconViewId int           `bson:"iconViewId"`
}

func (u *UseSelfItem) DumpDB() *UseSelfItemDumpDB {
	return &UseSelfItemDumpDB{
		Id:         u.bsonId,
		Name:       u.name,
		AgeisName:  u.ageisName,
		IconViewId: u.iconViewId,
	}
}

func (u *UseSelfItemDumpDB) Load() *UseSelfItem {
	return &UseSelfItem{
		Item: &Item{
			bsonId:     u.Id,
			name:       u.Name,
			ageisName:  u.AgeisName,
			iconViewId: u.IconViewId,
		},
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
