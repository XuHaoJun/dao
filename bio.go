package dao

type Bioer interface {
	Name() string
	Id() int
	SetId(int)
	DoJob(func())
	Run()
	ShutDown()
}

type BattleBioer interface {
	BattleInfo() BattleInfo
	IsDied() bool
	// Equip(item *Item)
	// UnEquip(itme)
	Level() int
	Str() int
	SetStr(int)
	Vit() int
	SetVit(int)
	Wis() int
	SetWis(int)
	Spi() int
	SetSpi(int)
	Def() int
	SetDef() int
	Mdef() int
	SetMdef() int
	Atk() int
	SetAtk() int
	Matk() int
	DecHp(int)
	SetMatk() int
	Hp() int
	MaxHp() int
	SetMaxHp() int
	Mp()
	MaxMp() int
	DecMp(int)
	SetMaxMp() int
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

// BattleBioBase imple Bioer, SceneBioer, BattleBioer
type BattleBioBase struct {
	*BioBase
	level  int
	isDied bool
	str    int
	vit    int
	wis    int
	spi    int
	def    int
	mdef   int
	atk    int
	matk   int
	maxHp  int
	hp     int
	maxMp  int
	mp     int
}

func NewBattleBioBase() *BattleBioBase {
	b := &BattleBioBase{
		BioBase: NewBioBase(),
		level:   1,
		isDied:  false,
		str:     0,
		vit:     0,
		wis:     0,
		spi:     0,
		def:     0,
		mdef:    0,
		atk:     0,
		matk:    0,
		maxHp:   0,
		hp:      0,
		maxMp:   0,
		mp:      0,
	}
	// TODO
	// imple BattleBioBase
	return b
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
