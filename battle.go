package dao

import (
	"github.com/xuhaojun/chipmunk"
)

type BattleInfo struct {
	isDied bool
	body   *chipmunk.Body
	level  int
	hp     int
	maxHp  int
	mp     int
	maxMp  int
	str    int
	vit    int
	wis    int
	spi    int
	atk    int
	matk   int
	def    int
	mdef   int
}

type BattleDamage struct {
	normal    int
	fire      int
	ice       int
	lightning int
	poison    int
}

type BattleDef struct {
	def                 int
	mdef                int
	fireResistance      int
	iceResistance       int
	lightningResistance int
	poisonResistance    int
}

func batttleAttrSub(damage int, def int) int {
	if def < damage && damage > 0 {
		damage -= def
	} else if def > damage && damage > 0 {
		damage = 0
	}
	return damage
}

func (bDamage *BattleDamage) SubBattleDef(bDef *BattleDef) *BattleDamage {
	bDamage.normal = batttleAttrSub(bDamage.normal, bDef.def)
	bDamage.fire = batttleAttrSub(bDamage.fire, bDef.fireResistance)
	bDamage.ice = batttleAttrSub(bDamage.ice, bDef.iceResistance)
	bDamage.lightning = batttleAttrSub(bDamage.lightning, bDef.lightningResistance)
	bDamage.poison = batttleAttrSub(bDamage.poison, bDef.poisonResistance)
	return bDamage
}

func (bDamage *BattleDamage) Total() int {
	return bDamage.fire + bDamage.ice + bDamage.normal + bDamage.lightning + bDamage.poison
}
