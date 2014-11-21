package dao

import (
	"fmt"
	"github.com/xuhaojun/chipmunk"
	"github.com/xuhaojun/chipmunk/vect"
	"math"
	"time"
)

type FireBallSkill struct {
	level            int
	ballLayer        chipmunk.Layer
	owner            Bioer
	fireBalls        map[int]*FireBallState
	afterUseDuration time.Duration
	delayDuration    time.Duration
}

func NewFireBallSkill(b Bioer) *FireBallSkill {
	return &FireBallSkill{
		level:         1,
		owner:         b,
		ballLayer:     -1,
		delayDuration: time.Second * 1,
		fireBalls:     make(map[int]*FireBallState),
	}
}

func (fbSkill *FireBallSkill) CanUse() bool {
	return fbSkill.afterUseDuration >= fbSkill.delayDuration
}

func (fbSkill *FireBallSkill) Fire() {
	if fbSkill.CanUse() {
		fbSkill.NewFireBallState().Fire()
		fbSkill.afterUseDuration = 0
	}

}

func (fbSkill *FireBallSkill) Update(delta float32) {
	deltaDuration := time.Duration(delta * float32(time.Second))
	fbSkill.afterUseDuration += deltaDuration
	for _, fb := range fbSkill.fireBalls {
		fb.Update(delta)
	}
}

// TODO
// replace simple int to dmage struct for descript more type damage!
func (f *FireBallSkill) BattleDamage() *BattleDamage {
	owner := f.owner
	minDmage := owner.Matk() * 1
	maxDmage := owner.Matk() * 3
	fireDamage := RandIntnRange(minDmage, maxDmage)
	fireDamage += f.level
	damage := &BattleDamage{
		fire: fireDamage,
	}
	return damage
}

type FireBallState struct {
	*SceneObject
	skill              *FireBallSkill
	battleDamage       *BattleDamage
	baseId             int
	inSceneDuration    time.Duration
	autoRemoveDuration time.Duration
	owner              Bioer
	hitCount           int
	maxHitCount        int
	bodyViewId         int
	iconViewId         int
	baseSpeed          vect.Float
	isInScene          bool
}

type FireBallStateClient struct {
	Id         int           `json:"id"`
	CpBody     *CpBodyClient `json:"cpBody"`
	BodyViewId int           `json:"bodyViewId"`
	IconViewId int           `json:"iconViewId"`
}

func (fbSkill *FireBallSkill) NewFireBallState() *FireBallState {
	fBall := &FireBallState{
		skill:              fbSkill,
		SceneObject:        &SceneObject{},
		baseId:             1,
		battleDamage:       fbSkill.BattleDamage(),
		inSceneDuration:    0,
		autoRemoveDuration: time.Second * 3,
		owner:              fbSkill.owner,
		hitCount:           0,
		maxHitCount:        1,
		bodyViewId:         10002,
		iconViewId:         1,
		baseSpeed:          100.0,
		isInScene:          false,
	}
	circle := chipmunk.NewCircle(vect.Vector_Zero, 9)
	circle.Layer = fbSkill.ballLayer
	circle.IsSensor = true
	body := chipmunk.NewBody(1, 1)
	body.AddShape(circle)
	body.UserData = fBall
	body.CallbackHandler = fBall
	fBall.body = body
	return fBall
}

func (f *FireBallState) SceneObjecter() SceneObjecter {
	return f
}

func (f *FireBallState) Client() interface{} {
	return &FireBallStateClient{
		Id:         f.id,
		CpBody:     ToCpBodyClient(f.body),
		BodyViewId: f.bodyViewId,
		IconViewId: f.iconViewId,
	}
}

func (f *FireBallState) Fire() {
	if f.isInScene || f.owner == nil {
		return
	}
	f.hitCount = 0
	f.inSceneDuration = 0
	f.isInScene = true
	body := f.body
	b := f.owner
	b.Scene().Add(f.SceneObjecter())
	body.SetAngle(b.Body().Angle())
	angle := b.Body().Angle() * -1
	bPos := b.Body().Position()
	dx := vect.Float(math.Sin(float64(angle)))
	dy := vect.Float(math.Cos(float64(angle)))
	bPos.Add(vect.Vect{X: 50 * dx, Y: 50 * dy})
	body.SetPosition(bPos)
	impulse := vect.Vect{X: dx, Y: dy}
	impulse.Mult(f.baseSpeed)
	// logger := b.World().logger
	// logger.Println("impulse: ", impulse)
	// body.AddForce(float32(impulse.X), float32(impulse.Y))
	body.SetVelocity(float32(impulse.X), float32(impulse.Y))
	f.skill.fireBalls[f.id] = f
}

func (f *FireBallState) Body() *chipmunk.Body {
	return f.body
}

func (f *FireBallState) Update(delta float32) {
	if !f.isInScene || f.scene == nil {
		return
	}
	if f.hitCount >= f.maxHitCount || f.inSceneDuration >= f.autoRemoveDuration {
		delete(f.skill.fireBalls, f.id)
		scene := f.scene
		scene.Remove(f.SceneObjecter())
		f.isInScene = false
		f.hitCount = 0
		f.inSceneDuration = 0
		return
	}
	// logger := f.owner.World().logger
	// logger.Println("fire ball pos: ", f.body.Position())
	// logger.Println("fire ball inSceneDuration: ", f.inSceneDuration)
	deltaDuration := time.Duration(delta * float32(time.Second))
	f.inSceneDuration += deltaDuration
}

func (f *FireBallState) HitTarget(b Bioer) {
	if f.owner == nil || b.IsDied() {
		return
	}
	b.TakeDamage(f.battleDamage, f.owner)
	fmt.Println("hit target: ", b.Name(), "damage: ", f.battleDamage, "b.hp: ", b.Hp())
}

func (f *FireBallState) OnCollideBioer(b Bioer) {
	if b == f.owner {
		return
	}
	f.HitTarget(b)
	f.hitCount += 1
	fmt.Println("hit bio:\n" + b.String())
}

func (f *FireBallState) CollisionEnter(arbiter *chipmunk.Arbiter) bool {
	b, ok := arbiter.BodyA.UserData.(Bioer)
	if ok {
		f.OnCollideBioer(b)
	}
	b, ok = arbiter.BodyB.UserData.(Bioer)
	if ok {
		f.OnCollideBioer(b)
	}
	return true
}

func (f *FireBallState) CollisionExit(arbiter *chipmunk.Arbiter) {
}

func (f *FireBallState) CollisionPreSolve(arbiter *chipmunk.Arbiter) bool {
	return true
}

func (f *FireBallState) CollisionPostSolve(arbiter *chipmunk.Arbiter) {}
