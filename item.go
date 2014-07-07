package dao

import (
	"labix.org/v2/mgo/bson"
)

type Item struct {
	id        int
	bsonId    bson.ObjectId
	name      string
	ageisName string
	// may be construct a view struct view.iconId or view.equipId
	iconViewId  int
	equipViewId int
}

type Equipment struct {
	*Item
	bonus      *BonusInfo
	etype      int
	equipLimit *EquipLimit
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

type EquipLimit struct {
	level int
	str   int
	vit   int
	wis   int
	spi   int
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
	Equip       map[int]*Equipment
	EtcItem     map[int]*EtcItem
	UseSelfItem map[int]*UseSelfItem
}
