package dao

import (
	"fmt"
)

type Pos struct {
	x float32
	y float32
}

type Scene struct {
	name         string
	mobs         map[int]*Mob
	npcs         map[int]*Npc
	chars        map[int]*Char
	etcItems     map[int]*EtcItem
	equipments   map[int]*Equipment
	useSelfItems map[int]*UseSelfItem
	job          chan func()
	quit         chan struct{}
}

type SceneInfo struct {
	Name string
	X    float32
	Y    float32
}

func NewScene(name string) *Scene {
	return &Scene{
		name:  name,
		mobs:  make(map[int]*Mob),
		npcs:  make(map[int]*Npc),
		chars: make(map[int]*Char),
		job:   make(chan func(), 1024),
		quit:  make(chan struct{}, 1),
	}
}

func (s *Scene) Run() {
	for {
		select {
		case job, ok := <-s.job:
			if !ok {
				return
			}
			job()
		case <-s.quit:
			close(s.job)
			s.quit <- struct{}{}
			return
		}
	}
}

func (s *Scene) ShutDown() {
	s.quit <- struct{}{}
	<-s.quit
}

func (s *Scene) DoJob(job func()) (err error) {
	defer handleErrSendCloseChanel(&err)
	s.job <- job
	return
}

func (s *Scene) FindMobById(mid int) *Mob {
	mobC := make(chan *Mob, 1)
	err := s.DoJob(func() {
		mob, ok := s.mobs[mid]
		if !ok || mob == nil {
			close(mobC)
		}
		mobC <- mob
	})
	if err != nil {
		close(mobC)
		return nil
	}
	mob, ok := <-mobC
	if !ok {
		return nil
	}
	return mob
}

func (s *Scene) DeleteBio(b SceneBioer) {
	s.DoJob(func() {
		switch b.(type) {
		case *Mob:
			mob := b.(*Mob)
			mid := mob.id
			s.mobs[mid] = nil
			mob.SetId(0)
			mob.SetScene(nil)
		case *Npc:
			npc := b.(*Npc)
			nid := npc.id
			s.npcs[nid] = nil
			npc.SetId(0)
			npc.SetScene(nil)
		case *Char:
			char := b.(*Char)
			cid := char.id
			s.chars[cid] = nil
			char.SetId(0)
			char.SetScene(nil)
		default:
			fmt.Println("you should never look this line.")
		}
	})
}

func (s *Scene) AddBio(b SceneBioer) {
	s.DoJob(func() {
		switch b.(type) {
		case *Mob:
			mob := b.(*Mob)
			mid := len(s.mobs)
			s.mobs[mid] = mob
			mob.SetId(mid)
			mob.SetScene(s)
		case *Npc:
			npc := b.(*Npc)
			nid := len(s.npcs)
			s.npcs[nid] = npc
			npc.SetId(nid)
			npc.SetScene(s)
		case *Char:
			char := b.(*Char)
			cid := len(s.chars)
			s.chars[cid] = char
			char.SetId(cid)
			char.SetScene(s)
		default:
			fmt.Println("you should never look this line.")
		}
	})
}

func (s *Scene) AddItem(i Itemer) {
}
