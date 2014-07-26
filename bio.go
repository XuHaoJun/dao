package dao

import (
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
	Move(x float32, y float32)
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
	viewAOI       *CircleAOI
	// base
	job  chan func()
	quit chan struct{}
}

type MoveState struct {
	moveCheckFunc func(skilCheckRunning bool) bool
	running       bool
	quit          chan struct{}
}

func NewBioBase() *BioBase {
	body := chipmunk.NewBody(1, 1)
	circle := chipmunk.NewCircle(vect.Vector_Zero, float32(32.0))
	body.SetPosition(vect.Vector_Zero)
	body.AddShape(circle)
	body.IgnoreGravity = true
	bio := &BioBase{
		enableViewAOI: true,
		name:          "",
		bodyViewId:    0,
		body:          body,
		scene:         nil,
		moveState: &MoveState{
			running:       false,
			moveCheckFunc: nil,
			quit:          make(chan struct{}, 1),
		},
		viewAOI: NewCircleAOI(160),
		job:     make(chan func(), 256),
		quit:    make(chan struct{}, 1),
	}
	bio.moveState.moveCheckFunc = bio.MoveCheckFunc()
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

func (b *BioBase) MoveCheckFunc() func(bool) bool {
	return func(skipCheckRunning bool) bool {
		tmpRunning := (b.moveState.running == true)
		if skipCheckRunning == true {
			tmpRunning = false
		}
		if b.scene == nil ||
			tmpRunning == true {
			return false
		}
		b.moveState.running = true
		return true
	}
}

func (b *BioBase) Move(x float32, y float32) {
	moveCheckC := make(chan bool, 1)
	err := b.DoJob(func() {
		moveCheckC <- b.moveState.moveCheckFunc(false)
	})
	if err != nil || <-moveCheckC == false {
		close(moveCheckC)
		return
	}
	timeC := time.Tick((1.0 * time.Second / 60.0)) // 60 fps
	defer func() {
		b.DoJob(func() {
			b.moveState.running = false
		})
	}()
	for {
		select {
		case <-timeC:
			err := b.DoJob(func() {
				moveCheck := b.moveState.moveCheckFunc(true)
				if moveCheck == false {
					moveCheckC <- false
					close(moveCheckC)
					return
				}
				// TODO
				// imple move
				moveCheckC <- true
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

func (b *BioBase) RunViewAOI() {
	// TODO
	// check b is in scene before run
	timeC := time.Tick(100.0 * time.Millisecond)
	for {
		select {
		case <-timeC:
			err := b.DoJob(func() {
				bioers := b.scene.Bioers()
				if bioers == nil {
					return
				}
				bioerBodys := make([]*chipmunk.Body, 0)
				for _, bioer := range bioers {
					body := bioer.Body()
					if body == nil {
						continue
					}
					body.SetVelocity(0, 0)
					body.IgnoreGravity = true
					bioerBodys = append(bioerBodys, body)
				}
				// TODO
				// imple bioerBodys collide detect with viewaoi's body
				// space := chipmunk.NewSpace()
				// space.AddBody
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
