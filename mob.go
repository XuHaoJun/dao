package dao

import (
	"github.com/xuhaojun/chipmunk"
	"github.com/xuhaojun/chipmunk/vect"
	"time"
)

var (
	MobLayer = chipmunk.Layer(4)
)

type Mober interface {
	Bioer
	MobClientBasic() *MobClientBasic
	RebornState() *MobRebornState
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

func (mr *MobRebornState) SetPositionFloat63(x, y float64) {
	mr.position.X = vect.Float(x)
	mr.position.Y = vect.Float(y)
}

func (m *Mob) InitSceneName() string {
	return m.initSceneName
}

func (m *Mob) SetInitSceneName(s string) {
	m.initSceneName = s
}

func (m *Mob) RebornState() *MobRebornState {
	return m.reborn
}

type Mob struct {
	*Bio
	baseId          int
	dropItemBaseIds []int
	//
	initSceneName string
	// reborn
	reborn *MobRebornState
	//
	aiUpdate func(delta float32)
	//
}

func NewMob(w *World) *Mob {
	mob := &Mob{
		Bio:             NewBio(w),
		baseId:          -1,
		reborn:          &MobRebornState{},
		dropItemBaseIds: []int{},
	}
	for _, shape := range mob.body.Shapes {
		shape.Layer = shape.Layer | MobLayer
	}
	mob.bodyViewId = 10001
	mob.clientCallPublisher = mob
	mob.Bio.skillUser = mob
	mob.Bio.partyer = mob
	mob.body.UserData = mob
	mob.viewAOIState = NewBioViewAOIState(200, mob.Bio)
	mob.OnBeKilled = mob.OnBeKilledFunc()
	mob.Bio.beKilleder = mob.Bioer()
	mob.fireBallSkill.ballLayer = CharLayer
	return mob
}

func (m *Mob) OnBeKilledFunc() func(killer Bioer) {
	return func(killer Bioer) {
		if m.reborn.enable == false {
			return
		}
		m.world.SetTimeout(func() {
			m.world.BioReborn <- m.Bioer()
		}, m.reborn.delayDuration)
		m.DropItem()
	}
}

func (m *Mob) DropItem() {
	length := len(m.dropItemBaseIds)
	if length <= 0 {
		return
	}
	n := m.world.util.Rand(0, length-1)
	itemBaseId := m.dropItemBaseIds[n]
	item, err := m.world.NewItemByBaseId(itemBaseId)
	if err != nil {
		return
	}
	if m.scene == nil {
		item.Body().SetPosition(m.lastPosition)
		scene := m.world.FindSceneByName(m.lastSceneName)
		if scene == nil {
			return
		}
		scene.Add(item)
	} else {
		item.Body().SetPosition(m.body.Position())
		scene := m.scene
		scene.Add(item)
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
	m.Emit("willReborn", m)
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

func (m *Mob) AfterUpdate(delta float32) {
	m.Bio.AfterUpdate(delta)
	if m.aiUpdate != nil {
		m.aiUpdate(delta)
	}
}
