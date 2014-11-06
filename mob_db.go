package dao

import (
	"github.com/xuhaojun/chipmunk/vect"
	"time"
)

func NewMobByBaseId(w *World, id int) *Mob {
	mob := NewMob(w)
	mob.baseId = id
	switch id {
	case 1:
		mob.name = "kiki"
		mob.vit = 5
		mob.str = 5
		mob.wis = 5
		mob.spi = 5
		mob.CalcAttributes()
		mob.hp = mob.maxHp
		mob.initSceneName = "daoField01"
		mob.reborn.enable = true
		mob.reborn.sceneName = "daoField01"
		mob.reborn.position = vect.Vect{X: 100, Y: 100}
		mob.reborn.delayDuration = time.Second * 5
	}
	return mob
}
