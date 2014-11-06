package dao

import (
	"github.com/xuhaojun/chipmunk/vect"
	"time"
)

type Mober interface {
	Bioer
	Reborn()
	MobClientBasic() *MobClientBasic
}

type MobClientBasic struct {
	BioClient *BioClientBasic `json:"bioConfig"`
}

type MobRebornState struct {
	enable          bool
	sceneName       string
	position        vect.Vect
	delayDuration   time.Duration
	currentDuration time.Duration
}

type Mob struct {
	*Bio
	baseId          int
	dropItemBaseIds []int
	//
	initSceneName string
	// reborn
	reborn *MobRebornState
}

func NewMob(w *World) *Mob {
	mob := &Mob{
		Bio:    NewBio(w),
		baseId: -1,
		reborn: &MobRebornState{},
	}
	mob.bodyViewId = 10001
	mob.clientCallPublisher = mob
	mob.Bio.skillUser = mob
	mob.body.UserData = mob
	mob.viewAOIState = NewBioViewAOIState(200, mob.Bio)
	mob.OnBeKilled = mob.OnBeKilledFunc()
	return mob
}

func (m *Mob) OnBeKilledFunc() func(killer Bioer) {
	return func(killer Bioer) {
		if m.reborn.enable == false {
			return
		}
		go func(w *World, mob *Mob) {
			select {
			case <-time.After(m.reborn.delayDuration):
				w.MobReborn <- mob
			}
		}(m.world, m)
	}
}

func (m *Mob) Reborn() {
	if m.reborn.enable == false {
		return
	}
	reborn := m.reborn
	w := m.world
	scene := w.FindSceneByName(reborn.sceneName)
	if scene == nil {
		return
	}
	m.hp = m.maxHp
	m.SetPosition(float32(reborn.position.X), float32(reborn.position.Y))
	scene.Add(m)
}

func (m *Mob) Bioer() Bioer {
	return m
}

func (m *Mob) Mober() Mober {
	return m
}

func (m *Mob) SceneObjecter() SceneObjecter {
	return m
}

func (m *Mob) MobClientBasic() *MobClientBasic {
	return &MobClientBasic{
		BioClient: m.Bio.BioClientBasic(),
	}
}

func (m *Mob) PublishClientCall(cc *ClientCall) {
	m.scene.DispatchClientCall(m, cc)
}
