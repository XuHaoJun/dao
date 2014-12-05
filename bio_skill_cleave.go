package dao

import (
	"fmt"
	"github.com/xuhaojun/chipmunk"
	"github.com/xuhaojun/chipmunk/vect"
	"math"
	"time"
)

type CleaveSkill struct {
	level            int
	baseId           int
	layer            chipmunk.Layer
	owner            Bioer
	cleaves          map[int]*CleaveState
	afterUseDuration time.Duration
	delayDuration    time.Duration
}

func NewCleaveSkill(b Bioer) *CleaveSkill {
	return &CleaveSkill{
		level:         1,
		baseId:        2,
		owner:         b,
		layer:         -1,
		delayDuration: time.Millisecond * 600,
		cleaves:       make(map[int]*CleaveState),
	}
}

func (skill *CleaveSkill) BattleDamage() *BattleDamage {
	owner := skill.owner
	minDmage := owner.Atk() * 1
	maxDmage := owner.Atk() * 3
	normalDamage := RandIntnRange(minDmage, maxDmage)
	normalDamage += skill.level
	return &BattleDamage{
		normal: normalDamage,
	}
}

type CleaveState struct {
	*SceneObject
	skill        *CleaveSkill
	battleDamage *BattleDamage
	owner        Bioer
	isInScene    bool
	bodyViewId   int
}

type CleaveClient struct {
	Id         int           `json:"id"`
	CpBody     *CpBodyClient `json:"cpBody"`
	BodyViewId int           `json:"bodyViewId"`
}

func (skill *CleaveSkill) NewCleaveState() *CleaveState {
	cleave := &CleaveState{
		skill:        skill,
		SceneObject:  &SceneObject{},
		battleDamage: skill.BattleDamage(),
		bodyViewId:   10003,
		isInScene:    false,
	}
	shape := chipmunk.NewBox(vect.Vector_Zero, 170, 55)
	shape.Layer = skill.layer
	shape.IsSensor = true
	body := chipmunk.NewBody(1, 1)
	body.AddShape(shape)
	body.UserData = cleave
	body.CallbackHandler = cleave
	cleave.body = body
	return cleave
}

func (skill *CleaveSkill) CanUse() bool {
	return skill.afterUseDuration >= skill.delayDuration
}

func (skill *CleaveSkill) Fire() {
	if skill.CanUse() {
		skill.NewCleaveState().Fire()
		skill.afterUseDuration = 0
	}

}

func (skill *CleaveSkill) Update(delta float32) {
	deltaDuration := time.Duration(delta * float32(time.Second))
	skill.afterUseDuration += deltaDuration
	for _, c := range skill.cleaves {
		c.Update(delta)
	}
}

// TODO
// replace it to an attack animation play on owner
func (c *CleaveState) Client() interface{} {
	return &CleaveClient{
		Id:         c.id,
		CpBody:     ToCpBodyClient(c.body),
		BodyViewId: c.bodyViewId,
	}
}

func (c *CleaveState) SceneObjecter() SceneObjecter {
	return c
}

func (c *CleaveState) Fire() {
	if c.isInScene || c.skill.owner == nil {
		return
	}
	c.isInScene = true
	body := c.body
	b := c.skill.owner
	body.SetAngle(b.Body().Angle())
	angle := b.Body().Angle() * -1
	bPos := b.Body().Position()
	dx := vect.Float(math.Sin(float64(angle)))
	dy := vect.Float(math.Cos(float64(angle)))
	bPos.Add(vect.Vect{X: 60 * dx, Y: 60 * dy})
	body.SetPosition(bPos)
	c.skill.cleaves[c.id] = c
	b.Scene().Add(c.SceneObjecter())
}

func (c *CleaveState) Body() *chipmunk.Body {
	return c.body
}

func (c *CleaveState) Destroy() {
	delete(c.skill.cleaves, c.id)
	scene := c.scene
	scene.Remove(c.SceneObjecter())
	c.isInScene = false
}

func (c *CleaveState) Update(delta float32) {
	if !c.isInScene || c.scene == nil {
		return
	}
	c.Destroy()
}

func (c *CleaveState) HitTarget(b Bioer) {
	if c.skill.owner == nil || b.IsDied() {
		return
	}
	b.TakeDamage(*c.battleDamage, c.skill.owner)
	fmt.Println(c.skill.owner.Name(), "cleave hit target: ", b.Name(),
		"damage: ", c.battleDamage, "b.hp: ", b.Hp())
}

func (c *CleaveState) OnCollideBioer(b Bioer) {
	if b == c.skill.owner {
		return
	}
	c.HitTarget(b)
}

func (c *CleaveState) CollisionEnter(arbiter *chipmunk.Arbiter) bool {
	b, ok := arbiter.BodyA.UserData.(Bioer)
	if ok {
		c.OnCollideBioer(b)
	}
	b, ok = arbiter.BodyB.UserData.(Bioer)
	if ok {
		c.OnCollideBioer(b)
	}
	return true
}

func (c *CleaveState) CollisionExit(arbiter *chipmunk.Arbiter) {
}

func (c *CleaveState) CollisionPreSolve(arbiter *chipmunk.Arbiter) bool {
	return true
}

func (c *CleaveState) CollisionPostSolve(arbiter *chipmunk.Arbiter) {}
