package dao

import (
	"fmt"
	"math"
	"time"

	"github.com/xuhaojun/chipmunk"
	"github.com/xuhaojun/chipmunk/vect"
)

type Bioer interface {
	Name() string
	Id() int
	GetId() int
	SetId(int)
	DoJob(func()) error
	Run()
	ShutDown()
	Move(vect.Vect)
	ShutDownMove()
	SetMoveTo(vect.Vect)
	Body() *chipmunk.Body
	Scene() *Scene
	SetScene(*Scene)
	GetScene() *Scene
	SetIdAndScene(int, *Scene)
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
	lastPosition  vect.Vect
	lastAngle     vect.Float
	running       bool
	space         *chipmunk.Space
	wallBodys     []*chipmunk.Body
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
			quit:         make(chan struct{}),
		},
		enableViewAOI: true,
		viewAOIRadius: 160.0,
		job:           make(chan func(), 256),
		quit:          make(chan struct{}),
	}
	bio.moveState.moveCheckFunc = bio.MoveCheckFunc()
	bio.viewAOI = NewViewAOI(bio.viewAOIRadius)
	bio.viewAOI.viewAOICheckFunc = bio.ViewAOICheckFunc()
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

func (b *BioBase) GetId() int {
	return b.id
}

func (b *BioBase) GetScene() *Scene {
	return b.scene
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

func (b *BioBase) SetId(id int) {
	b.DoJob(func() {
		b.id = id
	})
}

func (b *BioBase) DoSetIdAndScene(id int, scene *Scene) {
	b.id = id
	b.scene = scene
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
		return true
	}
}

func (b *BioBase) MoveSecondCheckFunc() func() bool {
	return func() bool {
		space := chipmunk.NewSpace()
		clone := b.body.Clone()
		clone.SetVelocity(
			float32(b.moveState.baseVelocity.X),
			float32(b.moveState.baseVelocity.Y))
		moveVelocity := clone.Velocity()
		moveVect := vect.Sub(b.moveState.targetPos, clone.Position())
		if float32(math.Abs(float64((1.0/60.0)*(moveVelocity.X)))) >
			float32(math.Abs(float64(moveVect.X))) {
			moveVect.X = 0
		}
		if float32(math.Abs(float64((1.0/60.0)*(moveVelocity.Y)))) >
			float32(math.Abs(float64(moveVect.Y))) {
			moveVect.Y = 0
		}
		if moveVect.X < 0 {
			moveVelocity.X *= -1
		} else if moveVect.X == 0 {
			moveVelocity.X = 0
		} else {
			moveVelocity.X *= 1
		}
		if moveVect.Y < 0 {
			moveVelocity.Y *= -1
		} else if moveVect.Y == 0 {
			moveVelocity.Y = 0
		} else {
			moveVelocity.Y *= 1
		}
		clone.SetVelocity(
			float32(moveVelocity.X),
			float32(moveVelocity.Y))
		newPos := clone.Position()
		// newPos.X += (moveVelocity.X * 1.0 / 60.0)
		// newPos.Y += (moveVelocity.Y * 1.0 / 60.0)
		newPos.X += 5
		newPos.Y += 5
		clone.SetPosition(newPos)
		collision := &BioOnStaticCallbacks{false}
		clone.CallbackHandler = collision
		space.AddBody(clone)
		wallBodys := b.scene.WallBodys()
		for _, body := range wallBodys {
			body.CallbackHandler = collision
			space.AddBody(body)
		}
		space.Step(1.0 / 60.0)
		space.Step(1.0 / 60.0)
		space.Step(1.0 / 60.0)
		space.Step(1.0 / 60.0)
		space.Step(1.0 / 60.0)
		fmt.Println("second collision.isOverlap: ",
			collision.isOverlap)
		return !collision.isOverlap
	}
}

func (b *BioBase) SetMoveTo(pos vect.Vect) {
	b.DoJob(func() {
		b.moveState.targetPos = pos
	})
}

type BioOnStaticCallbacks struct {
	isOverlap bool
}

func (na *BioOnStaticCallbacks) CollisionEnter(arbiter *chipmunk.Arbiter) bool {
	na.isOverlap = true
	fmt.Println("wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww")
	fmt.Println(arbiter.BodyA.Position())
	fmt.Println(arbiter.BodyB.Position())
	fmt.Println("wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww")
	return true
}

func (na *BioOnStaticCallbacks) CollisionPreSolve(arbiter *chipmunk.Arbiter) bool {
	return true
}

func (na *BioOnStaticCallbacks) CollisionPostSolve(arbiter *chipmunk.Arbiter) {}

func (na *BioOnStaticCallbacks) CollisionExit(arbiter *chipmunk.Arbiter) {}

func (b *BioBase) Move(pos vect.Vect) {
	moveCheckC := make(chan bool)
	err := b.DoJob(func() {
		b.moveState.targetPos = pos
		check := b.moveState.moveCheckFunc(false)
		if check == true {
			secondCheck := b.MoveSecondCheckFunc()()
			if secondCheck == false {
				moveCheckC <- false
				return
			}
			b.moveState.running = true
			b.moveState.space = chipmunk.NewSpace()
			b.moveState.wallBodys = b.scene.WallBodys()
			b.moveState.space.AddBody(b.body)
			for _, body := range b.moveState.wallBodys {
				b.moveState.space.AddBody(body)
			}
		}
		moveCheckC <- check
	})
	if err != nil || <-moveCheckC == false {
		close(moveCheckC)
		return
	}
	timeC := time.Tick((1.0 * time.Second / 60.0)) // 60 fps
	defer func() {
		b.DoJob(func() {
			b.moveState.running = false
			b.body = b.body.Clone()
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
				}
				space := b.moveState.space
				collisionOnStatic := &BioOnStaticCallbacks{false}
				b.body.CallbackHandler = collisionOnStatic
				// statics := b.scene.StaticBodys()
				// boxWall := chipmunk.b.body(1, 1)
				// wallTop := chipmunk.NewSegment(
				// 	vect.Vect{X: -100, Y: 70},
				// 	vect.Vect{X: 100, Y: 70}, 0)
				// wallTop.SetFriction(0)
				// wallTop.SetElasticity(0)
				// boxWall.AddShape(wallTop)
				// boxWall.CallbackHandler = collisionOnStatic
				// space.AddBody(boxWall)
				// TODO
				// 1. bio need lookat target x y
				// 2. send clientcall to chars on velocity or angle changed
				b.body.SetVelocity(
					float32(b.moveState.baseVelocity.X),
					float32(b.moveState.baseVelocity.Y))
				moveVelocity := b.body.Velocity()
				moveVect := vect.Sub(b.moveState.targetPos, b.body.Position())
				// check if this move one step > move vect directly reach it.
				if float32(math.Abs(float64((1.0/60.0)*(moveVelocity.X)))) >
					float32(math.Abs(float64(moveVect.X))) {
					reachX := vect.Vect{
						X: b.moveState.targetPos.X,
						Y: b.body.Position().Y,
					}
					b.body.SetPosition(reachX)
					moveVect.X = 0
				}
				if float32(math.Abs(float64((1.0/60.0)*(moveVelocity.Y)))) >
					float32(math.Abs(float64(moveVect.Y))) {
					reachY := vect.Vect{
						X: b.body.Position().X,
						Y: b.moveState.targetPos.Y,
					}
					b.body.SetPosition(reachY)
					moveVect.Y = 0
				}
				if vect.Equals(b.body.Position(), b.moveState.targetPos) ||
					vect.Equals(b.body.Velocity(), vect.Vector_Zero) {
					moveCheckC <- false
					return
				}
				if moveVect.X < 0 {
					moveVelocity.X *= -1
				} else if moveVect.X == 0 {
					moveVelocity.X = 0
				} else {
					moveVelocity.X *= 1
				}
				if moveVect.Y < 0 {
					moveVelocity.Y *= -1
				} else if moveVect.Y == 0 {
					moveVelocity.Y = 0
				} else {
					moveVelocity.Y *= 1
				}
				b.body.SetVelocity(
					float32(moveVelocity.X),
					float32(moveVelocity.Y))
				if vect.Equals(b.body.Position(), b.moveState.targetPos) ||
					vect.Equals(b.body.Velocity(), vect.Vector_Zero) {
					moveCheckC <- false
					return
				}
				b.body.LookAt(b.moveState.targetPos)
				b.moveState.lastPosition = b.body.Position()
				b.moveState.lastVelocity = b.body.Velocity()
				b.moveState.lastAngle = b.body.Angle()
				space.Step(1.0 / 60.0)
				// for _, body := range statics {
				// 	space.RemoveBody(body)
				// }
				// space.RemoveBody(boxWall)
				// space.RemoveBody(newBody)
				// TODO
				// publish charclient if velocity or angle change
				fmt.Println("b.body velocity X:", b.body.Velocity().X)
				fmt.Println("b.body velocity Y:", b.body.Velocity().Y)
				fmt.Println("b.body X:", b.body.Position().X)
				fmt.Println("b.body Y:", b.body.Position().Y)
				// space.Step(1.0 / 60.0)
				if vect.Equals(b.body.Position(), b.moveState.targetPos) ||
					vect.Equals(b.body.Velocity(), vect.Vector_Zero) ||
					collisionOnStatic.isOverlap == true {
					moveCheckC <- false
					return
				}
				b.body.CallbackHandler = nil
				moveCheckC <- true
			})
			if err != nil {
				return
			}
			mCheck, ok := <-moveCheckC
			if ok == false || mCheck == false {
				fmt.Println("wiwi")
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
	viewAOICheckFunc func(skipCheckRunning bool) bool
	OnBioEnterFunc   func(enter Bioer)
	OnBioLeaveFunc   func(leaver Bioer)
}

func NewViewAOI(r float32) *ViewAOI {
	aoi := &ViewAOI{NewCircleAOI(r), nil, nil, nil}
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

func (b *BioBase) ViewAOICheckFunc() func(bool) bool {
	return func(skipCheckRunning bool) bool {
		tmpRunning := (b.viewAOI.running == true)
		if skipCheckRunning == true {
			tmpRunning = false
		}
		if b.scene == nil ||
			tmpRunning == true {
			return false
		}
		b.viewAOI.running = true
		return true
	}
}

func (b *BioBase) RunViewAOI() {
	viewAOICheckC := make(chan bool, 1)
	err := b.DoJob(func() {
		viewAOICheckC <- b.viewAOI.viewAOICheckFunc(false)
	})
	if err != nil || <-viewAOICheckC == false {
		close(viewAOICheckC)
		return
	}
	timeC := time.Tick(100.0 * time.Millisecond)
	defer func() {
		b.DoJob(func() {
			b.viewAOI.running = false
		})
	}()
	for {
		select {
		case <-timeC:
			err := b.DoJob(func() {
				viewAOICheck := b.viewAOI.viewAOICheckFunc(true)
				if viewAOICheck == false {
					viewAOICheckC <- false
					return
				} else {
					viewAOICheckC <- true
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
			check, ok := <-viewAOICheckC
			if ok == false || check == false {
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
