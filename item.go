package dao

import (
	"labix.org/v2/mgo/bson"
)

type Item struct {
	id         int
	bsonId     bson.ObjectId
	name       string
	ageisName  string
	iconViewId int
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

type Equips struct {
	leftHand    *Equipment
	rightHand   *Equipment
	head        *Equipment
	shoulders   *Equipment
	torso       *Equipment
	wrists      *Equipment
	hands       *Equipment
	waist       *Equipment
	legs        *Equipment
	leftFinger  *Equipment
	rightFinger *Equipment
	neck        *Equipment
}

type EtcItem struct {
	*Item
}

type UseSelfItem struct {
	*Item
}

type Items struct {
	equipment   map[int]*Equipment
	etcItem     map[int]*EtcItem
	useSelfItem map[int]*UseSelfItem
}

type ItemsDumpDB struct {
	Equipment   map[string]*Equipment
	EtcItem     map[string]*EtcItem
	UseSelfItem map[string]*UseSelfItem
}

func NewItems() *Items {
	return &Items{
		make(map[int]*Equipment),
		make(map[int]*EtcItem),
		make(map[int]*UseSelfItem),
	}
}
