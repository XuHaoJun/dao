package dao

import (
	"fmt"
)

type Pos struct {
	x float32
	y float32
}

// TODO
// imple grid way
type Scene struct {
	name  string
	mobs  map[int]*Mob
	npcs  map[int]*Npc
	chars map[int]*Char
	job   chan func()
	quit  chan struct{}
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
			s.quit <- struct{}{}
			return
		}
	}
}

func (s *Scene) ShutDown() {
	s.quit <- struct{}{}
	<-s.quit
}

func (s *Scene) AddBio(b SceneBioer) {
	s.job <- func() {
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
			fmt.Println("you should look the line.")
		}
	}
}
