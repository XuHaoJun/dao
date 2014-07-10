package dao

import (
	"time"
)

type BattleBioer interface {
	// states
	BattleInfo() BattleInfo
	IsDied() bool
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
	// skills
	NormalAttack(b2 *BattleBioBase)
	// about items
	// Equip(item *Item)
	// UnEquip(itme)
}

// BattleBioBase imple Bioer, SceneBioer, BattleBioer
type BattleBioBase struct {
	*BioBase
	level             int
	isDied            bool
	str               int
	vit               int
	wis               int
	spi               int
	def               int
	mdef              int
	atk               int
	matk              int
	maxHp             int
	hp                int
	maxMp             int
	mp                int
	normalAttackState *NormalAttackState
}

type NormalAttackState struct {
	attackVelocity time.Duration
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
		attackVelocity: 2 * time.Second,
		running:        false,
		quit:           make(chan struct{}),
	}
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

func (b *BattleBioBase) BattleInfo() BattleInfo {
	battleC := make(chan BattleInfo, 1)
	b.job <- func() {
		battleC <- BattleInfo{
			isDied: b.isDied,
			pos:    b.Pos(),
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
	}
	return <-battleC
}

func (b *BattleBioBase) DecHp(n int) {
	b.job <- func() {
		if b.hp <= 0 {
			return
		}
		tmpHp := b.hp
		tmpHp -= n
		if tmpHp < 0 {
			b.hp = 0
			b.isDied = true
		} else {
			b.hp = tmpHp
		}
	}
}

func (b *BattleBioBase) NormalAttack(b2 *BattleBioBase) {
	attackVelocityC := make(chan time.Duration, 1)
	b.job <- func() {
		b.normalAttackState.running = true
		attackVelocityC <- b.normalAttackState.attackVelocity
	}
	timeC := time.Tick(<-attackVelocityC)
	for {
		select {
		case <-timeC:
			b.job <- func() {
				// TODO
				// add attack range check with
				// target's position
				target := b2.BattleInfo()
				if target.isDied == true {
					return
				}
				dmage := b.atk - target.def
				if dmage < 0 {
					dmage = 0
				}
				b2.DecHp(dmage)
				if b2.IsDied() {
					return
				}
			}
		case <-b.normalAttackState.quit:
			b.job <- func() {
				b.normalAttackState.running = false
			}
			return
		}
	}
}
