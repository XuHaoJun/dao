package dao

import (
	"time"

	"github.com/vova616/chipmunk"
	"github.com/vova616/chipmunk/vect"
)

// TODO
// imple AOI

type Bioer interface {
	Name() string
	Id() int
	SetId(int)
	DoJob(func()) error
	Run()
	ShutDown()
	Move(float32, float32)
}

type SceneBioer interface {
	Id() int
	SetId(int)
	Scene() *Scene
	SetScene(*Scene)
	X() float32
	SetX(float32)
	Y() float32
	SetY(float32)
	Pos() Pos
	SetPos(Pos)
}

// BioBase imple Bioer and SceneBioer
type BioBase struct {
	id         int
	name       string
	body       *chipmunk.Body
	bodyViewId int
	scene      *Scene
	moveState  *MoveState
	viewAOI    *CircleAOI
	job        chan func()
	quit       chan struct{}
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
		name:       "",
		bodyViewId: 0,
		body:       body,
		scene:      nil,
		moveState: &MoveState{
			running:       false,
			moveCheckFunc: nil,
			quit:          make(chan struct{}, 1),
		},
		viewAOI: &CircleAOI{
			radius:  128,
			body:    nil,
			running: false,
			quit:    make(chan struct{}, 1),
		},
		job:  make(chan func(), 512),
		quit: make(chan struct{}, 1),
	}
	bio.moveState.moveCheckFunc = bio.MoveCheckFunc()
	return bio
}

func (b *BioBase) Run() {
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

func (b *BioBase) Bioer() Bioer {
	return b
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

func (b *BioBase) X() float32 {
	xC := make(chan float32, 1)
	err := b.DoJob(func() {
		pos := b.body.Position()
		xC <- float32(pos.X)
	})
	if err != nil {
		close(xC)
		return 0
	}
	return <-xC
}

func (b *BioBase) SetX(x float32) {
	b.DoJob(func() {
		newPos := b.body.Position()
		newPos.X = vect.Float(x)
		b.body.SetPosition(newPos)
	})
}

func (b *BioBase) Y() float32 {
	yC := make(chan float32, 1)
	err := b.DoJob(func() {
		pos := b.body.Position()
		yC <- float32(pos.Y)
	})
	if err != nil {
		close(yC)
		return 0
	}
	return <-yC
}

func (b *BioBase) SetY(y float32) {
	b.DoJob(func() {
		newPos := b.body.Position()
		newPos.Y = vect.Float(y)
		b.body.SetPosition(newPos)
	})
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
	timeC := time.Tick((1 / 60) * time.Second) // 60 fps
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
					return
				}
				// TODO
				// imple move
				moveCheckC <- true
			})
			if err != nil || <-moveCheckC == false {
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
	radius  float32
	body    *chipmunk.Body
	bios    Bioer
	running bool
	quit    chan struct{}
}

func (b *BioBase) RunViewAOI() {
	timeC := time.Tick((1 / 60) * time.Second) // 60 fps
	for {
		select {
		case <-timeC:
			// TODO
			// 1. scan bios from scene like scene.Bios()
			// and update view bios.
			// 2. imple onenter and onleave callback for
			// bio enter or leave aoi
		case <-b.viewAOI.quit:
			return
		}
	}
}
