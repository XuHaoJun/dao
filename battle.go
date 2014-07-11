package dao

import (
	"github.com/vova616/chipmunk"
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
