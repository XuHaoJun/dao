package dao

import ()

type Pos struct {
	x float32
	y float32
}

// TODO:
// imple grid way
type Scene struct {
	name  string
	mobs  map[int]*Mob
	npcs  map[int]*Npc
	chars map[*Char]struct{}
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
		chars: make(map[*Char]struct{}),
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
			s.mobs[mob.id] = mob
		case *Npc:
			npc := b.(*Npc)
			s.npcs[npc.id] = npc
		case *Char:
			s.chars[b.(*Char)] = struct{}{}
		}
	}
}
