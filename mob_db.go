package dao

import (
	"github.com/xuhaojun/chipmunk/vect"
	"time"
)

func NewMobByBaseId(w *World, id int) Mober {
	var m Mober
	switch id {
	case 1:
		m = NewMobByBaseId1(w, 1)
	}
	return m
}

func NewMobByBaseId1(w *World, id int) Mober {
	mob := NewMob(w)
	mob.dropItemBaseIds = []int{1, 2, 5001, 5002, 5003, 10001}
	mob.viewAOIState = NewBioViewAOIState(350, mob.Bio)
	mob.baseId = id
	mob.name = "kiki"
	mob.level = 5
	mob.vit = 1
	mob.str = 1
	mob.wis = 1
	mob.spi = 1
	mob.CalcAttributes()
	mob.hp = mob.maxHp
	mob.initSceneName = "daoField01"
	mob.reborn.enable = true
	mob.reborn.sceneName = "daoField01"
	mob.reborn.position = vect.Vect{X: 350, Y: 350}
	mob.reborn.delayDuration = time.Second * 8
	mob.fireBallSkill.delayDuration = time.Second * 5
	mob.aiUpdate = func(delta float32) {
		sbs := mob.viewAOIState.inAreaSceneObjecters
		for sb, _ := range sbs {
			c, isCharer := sb.(Charer)
			if isCharer && mob.fireBallSkill.CanUse() {
				angle := float32(mob.SceneObject.body.Angle())
				mob.LookAtByBioer(c.(Bioer))
				newAngle := float32(mob.SceneObject.body.Angle())
				if newAngle != angle {
					clientCall := &ClientCall{
						Receiver: "bio",
						Method:   "handleUpdateCpBody",
						Params: []interface{}{
							mob.id,
							map[string]float32{
								"angle": float32(mob.SceneObject.body.Angle()),
							},
						},
					}
					mob.PublishClientCall(clientCall)
				}
				mob.UseFireBall()
				return
			}
		}
	}
	return mob
}
