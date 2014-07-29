package dao

import (
	"math"
	"time"

	"github.com/vova616/chipmunk"
	"github.com/vova616/chipmunk/vect"
)

type Bioer interface {
	Name() string
	Id() int
	SetId(int)
	DoJob(func()) error
	Run()
	ShutDown()
	Move(vect.Vect)
	Body() *chipmunk.Body
}

type SceneBioer interface {
	Id() int
	SetId(int)
	Scene() *Scene
	SetScene(*Scene)
	SetIdAndScene(int, *Scene)
	Body() *chipmunk.Body
}

// BioBase imple Bioer and SceneBioer
type BioBase struct {
	id         int
	name       string
	body       *chipmunk.Body
	bodyViewId int
	scene      *Scene
	moveState  *MoveState
	// aoi
	enableViewAOI bool
	viewAOIRadius float32
	viewAOI       *ViewAOI
	// base
	job  chan func()
	quit chan struct{}
}

type MoveState struct {
	moveCheckFunc func(skilCheckRunning bool) bool
	targetPos     vect.Vect
	baseVelocity  vect.Vect
	lastVelocity  vect.Vect
	lastAngle     vect.Float
	running       bool
	quit          chan struct{}
}

func NewBioBase() *BioBase {
	body := chipmunk.NewBody(1, 1)
	circle := chipmunk.NewCircle(vect.Vector_Zero, float32(32.0))
	circle.SetFriction(0)
	circle.SetElasticity(0)
	body.SetPosition(vect.Vector_Zero)
	body.SetVelocity(0, 0)
	body.SetMoment(chipmunk.Inf)
	body.AddShape(circle)
	body.IgnoreGravity = true
	bio := &BioBase{
		name:       "",
		bodyViewId: 0,
		body:       body,
		scene:      nil,
		moveState: &MoveState{
			running:      false,
			baseVelocity: vect.Vect{X: 10, Y: 10},
			quit:         make(chan struct{}, 1),
		},
		enableViewAOI: true,
		viewAOIRadius: 160.0,
		job:           make(chan func(), 256),
		quit:          make(chan struct{}, 1),
	}
	bio.moveState.moveCheckFunc = bio.MoveCheckFunc()
	bio.viewAOI = NewViewAOI(bio.viewAOIRadius)
	return bio
}

func (b *BioBase) Bioer() Bioer {
	return b
}

func (b *BioBase) Run() {
	if b.enableViewAOI == true {
		go b.RunViewAOI()
	}
	for {
		select {
		case job, ok := <-b.job:
			if !ok {
				return
			}
			job()
		case <-b.quit:
			close(b.job)
			if b.scene != nil {
				b.scene.DeleteBio(b)
				b.scene = nil
				b.id = 0
			}
			b.quit <- struct{}{}
			return
		}
	}
}

func (b *BioBase) DoJob(f func()) (err error) {
	defer handleErrSendCloseChanel(&err)
	b.job <- f
	return
}

func (b *BioBase) ShutDown() {
	b.quit <- struct{}{}
	<-b.quit
}

func (b *BioBase) Name() string {
	nameC := make(chan string, 1)
	err := b.DoJob(func() {
		nameC <- b.name
	})
	if err != nil {
		close(nameC)
		return ""
	}
	return <-nameC
}

func (b *BioBase) Scene() *Scene {
	sceneC := make(chan *Scene, 1)
	err := b.DoJob(func() {
		sceneC <- b.scene
	})
	if err != nil {
		close(sceneC)
		return nil
	}
	return <-sceneC
}

func (b *BioBase) SetScene(s *Scene) {
	b.DoJob(func() {
		b.scene = s
	})
}

func (b *BioBase) Body() *chipmunk.Body {
	bodyC := make(chan *chipmunk.Body, 1)
	err := b.DoJob(func() {
		bodyC <- b.body.Clone()
	})
	if err != nil {
		close(bodyC)
		return nil
	}
	return <-bodyC
}

func (b *BioBase) Pos() Pos {
	posC := make(chan Pos, 1)
	err := b.DoJob(func() {
		pos := b.body.Position()
		posC <- Pos{float32(pos.X), float32(pos.Y)}
	})
	if err != nil {
		close(posC)
		return Pos{}
	}
	return <-posC
}

func (b *BioBase) SetPos(p Pos) {
	b.DoJob(func() {
		x := vect.Float(p.X)
		y := vect.Float(p.Y)
		pos := vect.Vect{X: x, Y: y}
		b.body.SetPosition(pos)
	})
}

func (b *BioBase) SetId(id int) {
	b.DoJob(func() {
		b.id = id
	})
}

func (b *BioBase) SetIdAndScene(id int, scene *Scene) {
	b.DoJob(func() {
		b.id = id
		b.scene = scene
	})
}

func (b *BioBase) Id() int {
	idC := make(chan int, 1)
	err := b.DoJob(func() {
		idC <- b.id
	})
	if err != nil {
		close(idC)
		return 0
	}
	return <-idC
}

func (b *BioBase) ClientCallPublisher() ClientCallPublisher {
	return b
}

func (b *BioBase) PublishClientCall(c *ClientCall) {
	b.DoJob(func() {
		if b.scene == nil {
			return
		}
		b.scene.DispatchClientCall(b.ClientCallPublisher(), c)
	})
}

// FIXME
// may be check move target pos is same with bio's pos
func (b *BioBase) MoveCheckFunc() func(bool) bool {
	return func(skipCheckRunning bool) bool {
		tmpRunning := (b.moveState.running == true)
		if skipCheckRunning == true {
			tmpRunning = false
		}
		reached := vect.Equals(b.moveState.targetPos, b.body.Position())
		if reached == true ||
			b.scene == nil ||
			tmpRunning == true {
			return false
		}
		b.moveState.running = true
		return true
	}
}

func (b *BioBase) SetMoveTo(pos vect.Vect) {
	b.DoJob(func() {
		b.moveState.targetPos = pos
	})
}

func (b *BioBase) Move(pos vect.Vect) {
	moveCheckC := make(chan bool, 1)
	err := b.DoJob(func() {
		b.moveState.targetPos = pos
		if b.moveState.moveCheckFunc(false) == true {
			if b.moveState.targetPos == b.body.Position() {
				moveCheckC <- false
			} else {
				moveCheckC <- true
			}
		} else {
			moveCheckC <- false
		}
	})
	if err != nil || <-moveCheckC == false {
		close(moveCheckC)
		return
	}
	timeC := time.Tick((1.0 * time.Second / 60.0)) // 60 fps
	defer func() {
		b.DoJob(func() {
			b.moveState.running = false
			b.body.SetVelocity(0.0, 0.0)
		})
	}()
	for {
		select {
		case <-timeC:
			err := b.DoJob(func() {
				moveCheck := b.moveState.moveCheckFunc(true)
				if moveCheck == false {
					moveCheckC <- false
					return
				} else {
					moveCheckC <- true
				}
				statics := b.scene.StaticBodys()
				space := chipmunk.NewSpace()
				for _, body := range statics {
					space.AddBody(body)
				}
				space.AddBody(b.body)
				// TODO
				// 1. bio need lookat target x y
				// 2. send clientcall to chars on velocity or angle changed
				moveVelocity := b.moveState.baseVelocity
				moveVect := vect.Sub(b.moveState.targetPos, b.body.Position())
				// check if this move one step > move vect directly reach it.
				if math.Abs(float64((1.0/60.0)*(moveVelocity.X))) >
					math.Abs(float64(moveVect.X)) {
					reachX := vect.Vect{
						X: b.moveState.targetPos.X,
						Y: b.body.Position().Y,
					}
					b.body.SetPosition(reachX)
					moveVect.X = 0
				}
				if math.Abs(float64((1.0/60.0)*(moveVelocity.Y))) >
					math.Abs(float64(moveVect.Y)) {
					reachY := vect.Vect{
						X: b.body.Position().X,
						Y: b.moveState.targetPos.Y,
					}
					b.body.SetPosition(reachY)
					moveVect.Y = 0
				}
				if moveVect.X < 0 {
					moveVelocity.X *= -1
				} else if moveVect.X == 0 {
					moveVelocity.X = 0
				}
				if moveVect.Y < 0 {
					moveVelocity.Y *= -1
				} else if moveVect.Y == 0 {
					moveVelocity.Y = 0
				}
				b.moveState.lastVelocity = moveVelocity
				b.moveState.lastAngle = b.body.Angle()
				b.body.SetVelocity(float32(moveVect.X), float32(moveVect.Y))
				space.Step(1.0 / 60.0)
			})
			if err != nil {
				return
			}
			mCheck, ok := <-moveCheckC
			if ok == false || mCheck == false {
				return
			}
		case <-b.moveState.quit:
			return
		}
	}
}

func (b *BioBase) ShutDownMove() {
	b.DoJob(func() {
		if b.moveState.running {
			b.moveState.quit <- struct{}{}
		}
	})
}

type CircleAOI struct {
	running bool
	radius  float32
	body    *chipmunk.Body
	bioers  map[Bioer]struct{}
	quit    chan struct{}
}

func NewCircleAOI(r float32) *CircleAOI {
	body := chipmunk.NewBody(1, 1)
	circle := chipmunk.NewCircle(vect.Vector_Zero, r)
	circle.IsSensor = true
	body.SetPosition(vect.Vector_Zero)
	body.AddShape(circle)
	body.IgnoreGravity = true
	return &CircleAOI{
		running: false,
		radius:  r,
		body:    body,
		bioers:  make(map[Bioer]struct{}),
		quit:    make(chan struct{}, 1),
	}
}

type ViewAOI struct {
	*CircleAOI
	OnBioEnterFunc func(enter Bioer)
	OnBioLeaveFunc func(leaver Bioer)
}

func NewViewAOI(r float32) *ViewAOI {
	aoi := &ViewAOI{NewCircleAOI(r), nil, nil}
	return aoi
}

type ViewAOICallbacks struct {
	inAreaBioers map[Bioer]struct{}
}

func NewViewAOICallbacks() *ViewAOICallbacks {
	return &ViewAOICallbacks{
		make(map[Bioer]struct{}),
	}
}

func (v *ViewAOICallbacks) CollisionEnter(arbiter *chipmunk.Arbiter) bool {
	switch val := arbiter.BodyB.UserData.(type) {
	case Bioer:
		v.inAreaBioers[val] = struct{}{}
		// case Itemer:
	}
	return false
}

func (v *ViewAOICallbacks) CollisionPreSolve(arbiter *chipmunk.Arbiter) bool {
	return false
}

func (v *ViewAOICallbacks) CollisionPostSolve(arbiter *chipmunk.Arbiter) {}

func (v *ViewAOICallbacks) CollisionExit(arbiter *chipmunk.Arbiter) {}

func (b *BioBase) RunViewAOI() {
	// TODO
	// check b is in scene before run
	timeC := time.Tick(100.0 * time.Millisecond)
	for {
		select {
		case <-timeC:
			err := b.DoJob(func() {
				if b.scene == nil {
					return
				}
				bioers := b.scene.Bioers()
				if bioers == nil {
					return
				}
				space := chipmunk.NewSpace()
				b.viewAOI.body.CallbackHandler = NewViewAOICallbacks()
				space.AddBody(b.viewAOI.body)
				for _, bioer := range bioers {
					body := bioer.Body()
					if body == nil {
						continue
					}
					body.UserData = bioer
					body.SetVelocity(0, 0)
					body.IgnoreGravity = true
					space.AddBody(body)
				}
				space.Step(1)
				foundBioers := b.viewAOI.body.
					CallbackHandler.(*ViewAOICallbacks).
					inAreaBioers
				for bioer, _ := range b.viewAOI.bioers {
					_, in := foundBioers[bioer]
					if in == false {
						delete(b.viewAOI.bioers, bioer)
						if b.viewAOI.OnBioLeaveFunc != nil {
							b.viewAOI.OnBioLeaveFunc(bioer)
						}
					} else {
						delete(foundBioers, bioer)
					}
				}
				for bioer, _ := range foundBioers {
					b.viewAOI.bioers[bioer] = struct{}{}
					if b.viewAOI.OnBioEnterFunc != nil {
						b.viewAOI.OnBioEnterFunc(bioer)
					}
				}
			})
			if err != nil {
				return
			}
		case <-b.viewAOI.quit:
			return
		}
	}
}

func (b *BioBase) ShutDownViewAOI() {
	b.DoJob(func() {
		if b.viewAOI.running {
			b.viewAOI.quit <- struct{}{}
		}
	})
}
