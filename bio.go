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
	id         int
	name       string
	bodyViewId int
	pos        Pos
	scene      *Scene
	// equipItem map[*Item]struct{}
	// items map[]
	job  chan func()
	quit chan struct{}
}

func NewBioBase() *BioBase {
	return &BioBase{
		name:       "",
		bodyViewId: 0,
		pos:        Pos{0.0, 0.0},
		scene:      nil,
		job:        make(chan func(), 512),
		quit:       make(chan struct{}, 1),
	}
}

func (b *BioBase) Run() {
	b.job = make(chan func(), 512)
	b.quit = make(chan struct{}, 1)
	for {
		select {
		case job, ok := <-b.job:
			if !ok {
				return
			}
			job()
		case <-b.quit:
			close(b.job)
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

func (b *BioBase) X() float32 {
	xC := make(chan float32, 1)
	err := b.DoJob(func() {
		xC <- b.pos.x
	})
	if err != nil {
		close(xC)
		return 0
	}
	return <-xC
}

func (b *BioBase) SetX(x float32) {
	b.DoJob(func() {
		b.pos.x = x
	})
}

func (b *BioBase) Y() float32 {
	yC := make(chan float32, 1)
	err := b.DoJob(func() {
		yC <- b.pos.y
	})
	if err != nil {
		close(yC)
		return 0
	}
	return <-yC
}

func (b *BioBase) SetY(y float32) {
	b.DoJob(func() {
		b.pos.y = y
	})
}

func (b *BioBase) Pos() Pos {
	posC := make(chan Pos, 1)
	err := b.DoJob(func() {
		posC <- b.pos
	})
	if err != nil {
		close(posC)
		return Pos{}
	}
	return <-posC
}

func (b *BioBase) SetPos(p Pos) {
	b.DoJob(func() {
		b.pos = p
	})
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
