package dao

import ()

type Bioer interface {
	Name() string
	Id() int
	SetId(int)
	DoJob(func())
	Run()
	ShutDown()
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
	id    int
	name  string
	pos   Pos
	scene *Scene
	// equipItem map[*Item]struct{}
	// items map[]
	job  chan func()
	quit chan struct{}
}

func NewBioBase() *BioBase {
	return &BioBase{
		name:  "",
		pos:   Pos{0.0, 0.0},
		scene: nil,
		job:   make(chan func(), 512),
		quit:  make(chan struct{}, 1),
	}
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
			b.quit <- struct{}{}
			return
		}
	}
}

func (b *BioBase) DoJob(f func()) {
	b.job <- f
}

func (b *BioBase) ShutDown() {
	b.quit <- struct{}{}
	<-b.quit
}

func (b *BioBase) Name() string {
	nameC := make(chan string, 1)
	b.job <- func() {
		nameC <- b.name
	}
	return <-nameC
}

func (b *BioBase) Scene() *Scene {
	sceneC := make(chan *Scene, 1)
	b.job <- func() {
		sceneC <- b.scene
	}
	return <-sceneC
}

func (b *BioBase) SetScene(s *Scene) {
	b.job <- func() {
		b.scene = s
	}
}

func (b *BioBase) X() float32 {
	xC := make(chan float32, 1)
	b.job <- func() {
		xC <- b.pos.x
	}
	return <-xC
}

func (b *BioBase) SetX(x float32) {
	b.job <- func() {
		b.pos.x = x
	}
}

func (b *BioBase) Y() float32 {
	yC := make(chan float32, 1)
	b.job <- func() {
		yC <- b.pos.y
	}
	return <-yC
}

func (b *BioBase) SetY(y float32) {
	b.job <- func() {
		b.pos.y = y
	}
}

func (b *BioBase) Pos() Pos {
	posC := make(chan Pos, 1)
	b.job <- func() {
		posC <- b.pos
	}
	return <-posC
}

func (b *BioBase) SetPos(p Pos) {
	b.job <- func() {
		b.pos = p
	}
}

func (b *BioBase) SetId(id int) {
	b.job <- func() {
		b.id = id
	}
}

func (b *BioBase) Id() int {
	idC := make(chan int, 1)
	b.job <- func() {
		idC <- b.id
	}
	return <-idC
}
