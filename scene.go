package dao

import (
	"fmt"
	"github.com/xuhaojun/chipmunk"
	"github.com/xuhaojun/chipmunk/vect"
)

type Pos struct {
	X float32
	Y float32
}

type SceneBios struct {
	mobs  map[int]*Mob
	npcs  map[int]*Npc
	chars map[int]*Char
}

type Scene struct {
	name string
	// bios and items
	mobs  map[int]*Mob
	npcs  map[int]*Npc
	chars map[int]*Char
	items map[int]Itemer
	//
	wallBodys   []*chipmunk.Body
	staticBodys map[*chipmunk.Body]struct{}
	//
	job  chan func()
	quit chan struct{}
}

type SceneBodyer interface {
	Body() *chipmunk.Body
}

type SceneInfo struct {
	Name string
	X    float32
	Y    float32
}

func NewScene(name string) *Scene {
	return &Scene{
		name:        name,
		mobs:        make(map[int]*Mob),
		npcs:        make(map[int]*Npc),
		chars:       make(map[int]*Char),
		items:       make(map[int]Itemer),
		wallBodys:   make([]*chipmunk.Body, 0),
		staticBodys: make(map[*chipmunk.Body]struct{}),
		job:         make(chan func(), 1024),
		quit:        make(chan struct{}, 1),
	}
}

func NewWallScene(name string, w vect.Float, h vect.Float) *Scene {
	s := NewScene(name)
	boxWall := chipmunk.NewBodyStatic()
	wallTop := chipmunk.NewSegment(vect.Vect{X: -w / 2, Y: h / 2}, vect.Vect{X: w / 2, Y: h / 2}, 0)
	wallTop.SetFriction(0)
	wallTop.SetElasticity(0)
	boxWall.AddShape(wallTop)
	wallBottom := chipmunk.NewSegment(vect.Vect{X: -w / 2, Y: -h / 2}, vect.Vect{X: w / 2, Y: -h / 2}, 0)
	wallBottom.SetFriction(0)
	wallBottom.SetElasticity(0)
	boxWall.AddShape(wallBottom)
	wallLeft := chipmunk.NewSegment(vect.Vect{X: -w / 2, Y: h / 2}, vect.Vect{X: -w / 2, Y: -h / 2}, 0)
	wallLeft.SetFriction(0)
	wallLeft.SetElasticity(0)
	boxWall.AddShape(wallLeft)
	wallRight := chipmunk.NewSegment(vect.Vect{X: w / 2, Y: h / 2}, vect.Vect{X: w / 2, Y: -h / 2}, 0)
	wallRight.SetFriction(0)
	wallRight.SetElasticity(0)
	boxWall.AddShape(wallRight)
	s.wallBodys = append(s.wallBodys, boxWall)
	s.staticBodys[boxWall] = struct{}{}
	return s
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

// func (s *Scene) StaticBody() map[*chipmunk.Body]struct{} {
// 	sbodysC := make(chan map[*chipmunk.Body]struct{}, 1)
// 	err := s.DoJob(func() {
// 		<-s.staticBody
// 	})
// 	if err != nil {
// 		close(sbodysC)
// 		return nil
// 	}
// 	return nil
// }

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

func (s *Scene) DeleteBio(b Bioer) {
	s.DoJob(func() {
		fmt.Println("delete bioing start")
		switch b.(type) {
		case *Mob:
			var mid int
			m := b.(*Mob)
			for id, mob := range s.mobs {
				if mob == m {
					mid = id
					break
				}
			}
			s.mobs[mid] = nil
		case *Npc:
			var nid int
			n := b.(*Npc)
			for id, npc := range s.npcs {
				if npc == n {
					nid = id
					break
				}
			}
			s.npcs[nid] = nil
		case *Char:
			var cid int
			c := b.(*Char)
			for id, char := range s.chars {
				if char == c {
					cid = id
					break
				}
			}
			s.chars[cid] = nil
		default:
			fmt.Println("you should never look this line.")
			return
		}
		fmt.Println("delete bioing done!")
	})
}

func (s *Scene) Bioers() []Bioer {
	bioersC := make(chan []Bioer, 1)
	err := s.DoJob(func() {
		bioers := make([]Bioer, 0)
		for _, mob := range s.mobs {
			if mob != nil {
				bioers = append(bioers, mob.Bioer())
			}
		}
		for _, npc := range s.npcs {
			if npc != nil {
				bioers = append(bioers, npc.Bioer())
			}
		}
		for _, char := range s.chars {
			if char != nil {
				bioers = append(bioers, char.Bioer())
			}
		}
		bioersC <- bioers
	})
	if err != nil {
		close(bioersC)
		return nil
	}
	return <-bioersC
}

func (s *Scene) SceneBios() *SceneBios {
	biosC := make(chan *SceneBios, 1)
	err := s.DoJob(func() {
		mobs := make(map[int]*Mob, len(s.mobs))
		for i, mob := range s.mobs {
			if mob != nil {
				mobs[i] = mob
			}
		}
		npcs := make(map[int]*Npc, len(s.npcs))
		for i, npc := range s.npcs {
			if npc != nil {
				npcs[i] = npc
			}
		}
		chars := make(map[int]*Char, len(s.chars))
		for i, char := range s.chars {
			if char != nil {
				chars[i] = char
			}
		}
		bios := &SceneBios{
			mobs:  mobs,
			npcs:  npcs,
			chars: chars,
		}
		biosC <- bios
	})
	if err != nil {
		close(biosC)
		return nil
	}
	return <-biosC
}

func (s *Scene) AddBio(b Bioer) {
	s.DoJob(func() {
		fmt.Println(b.Name())
		doneC := make(chan struct{}, 1)
		err := b.DoJob(func() {
			if b.GetScene() == s {
				return
			}
			switch b.(type) {
			case *Mob:
				mob := b.(*Mob)
				mid := len(s.mobs) + 1
				s.mobs[mid] = mob
				mob.DoSetIdAndScene(mid, s)
			case *Npc:
				npc := b.(*Npc)
				nid := len(s.npcs) + 1
				s.npcs[nid] = npc
				npc.DoSetIdAndScene(nid, s)
			case *Char:
				char := b.(*Char)
				cid := len(s.chars) + 1
				s.chars[cid] = char
				char.DoSetIdAndScene(cid, s)
			default:
				fmt.Println("you should never look this line.")
			}
			doneC <- struct{}{}
		})
		if err != nil {
			return
		}
		<-doneC
	})
}

func (s *Scene) WallBodys() []*chipmunk.Body {
	sbodysC := make(chan []*chipmunk.Body, 1)
	err := s.DoJob(func() {
		sbodys := make([]*chipmunk.Body, len(s.wallBodys))
		for i, body := range s.wallBodys {
			sbodys[i] = body.Clone()
		}
		sbodysC <- sbodys
	})
	if err != nil {
		close(sbodysC)
		return nil
	}
	return <-sbodysC
}

func (s *Scene) StaticBodys() []*chipmunk.Body {
	sbodysC := make(chan []*chipmunk.Body, 1)
	err := s.DoJob(func() {
		sbodys := make([]*chipmunk.Body, len(s.staticBodys))
		i := 0
		for body, _ := range s.staticBodys {
			sbodys[i] = body
			i++
		}
		sbodysC <- sbodys
	})
	if err != nil {
		close(sbodysC)
		return nil
	}
	return <-sbodysC
}

type ClientCallPublisher interface {
	PublishClientCall(c *ClientCall)
}

func (s *Scene) DispatchClientCall(sender ClientCallPublisher, c *ClientCall) {
	s.DoJob(func() {
		for _, char := range s.chars {
			if char == nil {
				continue
			}
			char.OnReceiveClientCall(sender, c)
		}
	})
}

func (s *Scene) DeleteItem(item Itemer) {
	s.DoJob(func() {
		for i, foundItem := range s.items {
			if foundItem == item {
				s.items[i] = nil
				return
			}
		}
	})
}

func (s *Scene) AddItem(i Itemer) {
	s.DoJob(func() {
		i.SetScene(s)
		s.items[len(s.items)] = i
	})
}

func (s *Scene) FindItemId(id int) Itemer {
	itemC := make(chan Itemer, 1)
	err := s.DoJob(func() {
		item, ok := s.items[id]
		if !ok {
			itemC <- nil
			return
		}
		itemC <- item
	})
	if err != nil {
		close(itemC)
		return nil
	}
	return <-itemC
}
