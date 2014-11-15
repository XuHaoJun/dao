package dao

import (
	"fmt"
	"github.com/xuhaojun/chipmunk"
	"github.com/xuhaojun/chipmunk/vect"
	"math"
	"time"
)

type FireBallState struct {
	*SceneObject
	baseId             int
	level              int
	inSceneDuration    time.Duration
	autoRemoveDuration time.Duration
	owner              Bioer
	hitCount           int
	maxHitCount        int
	bodyViewId         int
	iconViewId         int
	baseSpeed          vect.Float
	isFired            bool
}

type FireBallStateClient struct {
	Id         int           `json:"id"`
	CpBody     *CpBodyClient `json:"cpBody"`
	BodyViewId int           `json:"bodyViewId"`
	IconViewId int           `json:"iconViewId"`
}

func NewFireBallState(b Bioer) *FireBallState {
	fBall := &FireBallState{
		SceneObject:        &SceneObject{},
		baseId:             1,
		level:              1,
		inSceneDuration:    0,
		autoRemoveDuration: time.Second * 3,
		owner:              b,
		hitCount:           0,
		maxHitCount:        1,
		bodyViewId:         10002,
		iconViewId:         1,
		baseSpeed:          100.0,
		isFired:            false,
	}
	circle := chipmunk.NewCircle(vect.Vector_Zero, 6)
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
	if f.isFired {
		return
	}
	f.hitCount = 0
	f.inSceneDuration = 0
	f.isFired = true
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
}

func (f *FireBallState) CanUse() bool {
	return !f.isFired
}

func (f *FireBallState) Body() *chipmunk.Body {
	return f.body
}

func (f *FireBallState) Update(delta float32) {
	if !f.isFired {
		return
	}
	if f.hitCount >= f.maxHitCount || f.inSceneDuration >= f.autoRemoveDuration {
		scene := f.scene
		f.lastId = f.id
		f.lastSceneName = scene.name
		scene.Remove(f.SceneObjecter())
		f.isFired = false
		return
	}
	// logger := f.owner.World().logger
	// logger.Println("fire ball pos: ", f.body.Position())
	// logger.Println("fire ball inSceneDuration: ", f.inSceneDuration)
	deltaDuration := time.Duration(delta * float32(time.Second))
	f.inSceneDuration += deltaDuration
}

// TODO
// replace simple int to dmage struct for descript more type damage!
func (f *FireBallState) Damage() int {
	if f.owner == nil {
		return 0
	}
	owner := f.owner
	minDmage := owner.Wis() * 1
	maxDmage := owner.Wis() * 3
	dmage := RandIntnRange(minDmage, maxDmage)
	return dmage
}

func (f *FireBallState) HitTarget(b Bioer) {
	if b.IsDied() {
		return
	}
	b.TakeDamage(f.Damage(), f.owner)
	fmt.Println("hit target: ", b.Name(), "damage: ", f.Damage(), "b.hp: ", b.Hp())
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
