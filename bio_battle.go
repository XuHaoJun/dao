package dao

import (
	"time"

	"github.com/vova616/chipmunk"
	"github.com/vova616/chipmunk/vect"
)

type BattleBioer interface {
	// states
	BattleInfo() *BattleInfo
	Level() int
	IsDied() bool
	DecHp(int) bool
	DecMp(int)
	CalcAttributes()
	DoCalcAttributes()
	// callbacks
	OnBeKilledFunc() func()
	OnKillFunc() func(target BattleBioer)
	// skills
	NormalAttack(b2 BattleBioer)
	ShutDownNormalAttack()
}

// BattleBioBase imple Bioer, SceneBioer, BattleBioer
type BattleBioBase struct {
	*BioBase
	level  int
	isDied bool
	// main attribue
	str int
	vit int
	wis int
	spi int
	// sub attribue
	def   int
	mdef  int
	atk   int
	matk  int
	maxHp int
	hp    int
	maxMp int
	mp    int
	// callbacks
	OnKill     func(target BattleBioer)
	OnBeKilled func()
	// skills
	normalAttackState *NormalAttackState
}

type NormalAttackState struct {
	attackTimeStep time.Duration
	attackRadius   float32
	running        bool
	quit           chan struct{}
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
	// FIXME
	// may be not use 2 sec for slowest attack velocity
	b.normalAttackState = &NormalAttackState{
		attackTimeStep: 2 * time.Second,
		attackRadius:   2.0,
		running:        false,
		quit:           make(chan struct{}, 1),
	}
	b.OnKill = b.OnKillFunc()
	b.OnBeKilled = b.OnBeKilledFunc()
	return b
}

func (b *BattleBioBase) Run() {
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

func (b *BattleBioBase) DoCalcAttributes() {
	b.atk = b.str * 5
	b.maxHp = b.vit * 5
	b.def = b.vit * 3
	b.maxMp = b.wis * 5
	b.mdef = b.wis * 3
	if b.hp > b.maxHp {
		b.hp = b.maxHp
	}
	if b.mp > b.maxMp {
		b.mp = b.maxMp
	}
}

func (b *BattleBioBase) CalcAttributes() {
	b.DoJob(func() {
		b.DoCalcAttributes()
	})
}

func (b *BattleBioBase) BattleBioer() BattleBioer {
	return b
}

func (b *BattleBioBase) Level() int {
	levelC := make(chan int, 1)
	err := b.DoJob(func() {
		levelC <- b.level
	})
	if err != nil {
		close(levelC)
		return 0
	}
	return <-levelC
}

func (b *BattleBioBase) IsDied() bool {
	c := make(chan bool, 1)
	err := b.DoJob(func() {
		c <- b.isDied
	})
	if err != nil {
		close(c)
		return true
	}
	return <-c
}

func (b *BattleBioBase) BattleInfo() *BattleInfo {
	battleC := make(chan *BattleInfo, 1)
	err := b.DoJob(func() {
		battleC <- &BattleInfo{
			isDied: b.isDied,
			body:   b.body.Clone(),
			level:  b.level,
			hp:     b.hp,
			maxHp:  b.maxHp,
			mp:     b.mp,
			maxMp:  b.maxMp,
			str:    b.str,
			vit:    b.vit,
			wis:    b.wis,
			spi:    b.spi,
			atk:    b.atk,
			matk:   b.matk,
			def:    b.def,
			mdef:   b.mdef,
		}
	})
	if err != nil {
		close(battleC)
		return nil
	}
	return <-battleC
}

func (b *BattleBioBase) DecHp(n int) bool {
	killedC := make(chan bool, 1)
	err := b.DoJob(func() {
		if b.hp <= 0 {
			killedC <- false
			return
		}
		tmpHp := b.hp
		tmpHp -= n
		if tmpHp < 0 {
			b.hp = 0
			b.isDied = true
			f := b.OnBeKilledFunc()
			f()
			killedC <- true
		} else {
			b.hp = tmpHp
			killedC <- false
		}
	})
	if err != nil {
		close(killedC)
		return false
	}
	return <-killedC
}

func (b *BattleBioBase) DecMp(n int) {
	b.DoJob(func() {
		if b.mp <= 0 {
			return
		}
		tmpMp := b.mp
		tmpMp -= n
		if tmpMp < 0 {
			b.mp = 0
		} else {
			b.mp = tmpMp
		}
	})
}

func (b *BattleBioBase) OnKillFunc() func(target BattleBioer) {
	return func(target BattleBioer) {}
}

func (b *BattleBioBase) OnBeKilledFunc() func() {
	return func() {
		if b.scene != nil {
			b.scene.DeleteBio(b)
		}
	}
}

type NormalAttackCallbacks struct {
	isOverlap bool
}

func (na *NormalAttackCallbacks) CollisionEnter(arbiter *chipmunk.Arbiter) bool {
	na.isOverlap = true
	return false
}

func (na *NormalAttackCallbacks) CollisionPreSolve(arbiter *chipmunk.Arbiter) bool {
	return false
}

func (na *NormalAttackCallbacks) CollisionPostSolve(arbiter *chipmunk.Arbiter) {}

func (na *NormalAttackCallbacks) CollisionExit(arbiter *chipmunk.Arbiter) {}

func (b *BattleBioBase) NormalAttack(b2 BattleBioer) {
	if b.IsDied() || b2.IsDied() || b == b2 {
		return
	}
	attackTimeStepC := make(chan time.Duration, 1)
	err := b.DoJob(func() {
		b.normalAttackState.running = true
		// TODO
		// should detect attack radius from weapon
		attackTimeStepC <- b.normalAttackState.attackTimeStep
	})
	if err != nil {
		close(attackTimeStepC)
		return
	}
	timeC := time.Tick(<-attackTimeStepC)
	for {
		select {
		case <-timeC:
			b.DoJob(func() {
				// TODO
				// add attack radius check with
				// target's position
				target := b2.BattleInfo()
				if target.isDied == true {
					b.ShutDownNormalAttack()
					return
				}
				space := chipmunk.NewSpace()
				attackRange := chipmunk.NewBody(1, 1)
				rangeShape := chipmunk.NewCircle(
					vect.Vector_Zero,
					b.normalAttackState.attackRadius)
				rangeShape.IsSensor = true
				attackRange.AddShape(rangeShape)
				attackRange.SetPosition(b.body.Position())
				check := &NormalAttackCallbacks{false}
				attackRange.CallbackHandler = check
				space.AddBody(attackRange)
				space.AddBody(target.body)
				space.Step(0.1)
				if check.isOverlap == false {
					b.ShutDownNormalAttack()
					return
				}
				dmage := b.atk - target.def
				if dmage < 0 {
					dmage = 0
				}
				killed := b2.DecHp(dmage)
				if killed {
					f := b.OnKillFunc()
					f(b2)
					b.ShutDownNormalAttack()
					return
				}
				if b2.IsDied() {
					b.ShutDownNormalAttack()
					return
				}
			})
		case <-b.normalAttackState.quit:
			b.DoJob(func() {
				b.normalAttackState.running = false
			})
			return
		}
	}
}

func (b *BattleBioBase) ShutDownNormalAttack() {
	b.DoJob(func() {
		if b.normalAttackState.running {
			b.normalAttackState.quit <- struct{}{}
		}
	})
}
