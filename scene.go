package dao

import (
	"github.com/xuhaojun/chipmunk"
	"github.com/xuhaojun/chipmunk/vect"
	"time"
)

type Scene struct {
	name   string
	width  float32
	height float32
	//
	idCounter int
	//
	world *World
	//
	sceneObjects map[int]SceneObjecter
	//
	staticBodys map[*chipmunk.Body]struct{}
	//
	cpSpace *chipmunk.Space
	//
	defaultGroundTextureName string
	//
	autoClearItemDuration time.Duration
}

type SceneInfo struct {
	Name string
	X    float32
	Y    float32
}

func NewScene(w *World, name string) *Scene {
	cpSpace := chipmunk.NewSpace()
	cpSpace.Iterations = 10
	return &Scene{
		world:        w,
		name:         name,
		idCounter:    1,
		sceneObjects: make(map[int]SceneObjecter),
		staticBodys:  make(map[*chipmunk.Body]struct{}),
		cpSpace:      cpSpace,
		//
		defaultGroundTextureName: "grass",
		//
		autoClearItemDuration: time.Second * 10,
	}
}

type SceneClient struct {
	Name        string          `json:"name"`
	StaticBodys []*CpBodyClient `json:"staticBodys"`
	Run         bool            `json:"run"`
	Width       float32         `json:"width"`
	Height      float32         `json:"height"`
	//
	DefaultGroundTextureName string `json:"defaultGroundTextureName"`
}

func (s *Scene) SceneClient() *SceneClient {
	i := 0
	cpBodyClients := make([]*CpBodyClient, len(s.staticBodys))
	for sbody, _ := range s.staticBodys {
		cpBodyClients[i] = ToCpBodyClient(sbody)
		i = i + 1
	}
	return &SceneClient{
		Name:        s.name,
		StaticBodys: cpBodyClients,
		Run:         false,
		Width:       s.width,
		Height:      s.height,
		DefaultGroundTextureName: s.defaultGroundTextureName,
	}
}

func NewWallScene(world *World, name string, w vect.Float, h vect.Float) *Scene {
	s := NewScene(world, name)
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
	s.cpSpace.AddBody(boxWall)
	s.staticBodys[boxWall] = struct{}{}
	s.width = float32(w)
	s.height = float32(h)
	return s
}

func (s *Scene) Update(delta float32) {
	for _, sb := range s.sceneObjects {
		sb.BeforeUpdate(delta)
	}
	s.cpSpace.Step(vect.Float(delta))
	for _, sb := range s.sceneObjects {
		sb.AfterUpdate(delta)
		deltaTime := time.Duration(float32(time.Second) * delta)
		sb.IncInSceneDuration(deltaTime)
		item, isItem := sb.(Itemer)
		if isItem &&
			item.InSceneDuration() >= s.autoClearItemDuration {
			s.Remove(item.SceneObjecter())
		}
	}
}

type ClientCallPublisher interface {
	PublishClientCall(c *ClientCall)
}

func (s *Scene) DispatchClientCall(sender ClientCallPublisher, c *ClientCall) {
	for _, sb := range s.sceneObjects {
		char, ok := sb.(Charer)
		if ok && char.Scene() != nil {
			char.OnReceiveClientCall(sender, c)
		}
	}
}

func (s *Scene) FindMobById(mid int) *Mob {
	mob, ok := s.sceneObjects[mid].(*Mob)
	if ok {
		return mob
	}
	return nil
}

func (s *Scene) FindNpcById(nid int) *Npc {
	npc, ok := s.sceneObjects[nid].(*Npc)
	if ok {
		return npc
	}
	return nil
}

func (s *Scene) FindItemerById(nid int) Itemer {
	if nid < 0 {
		return nil
	}
	itemer, ok := s.sceneObjects[nid].(Itemer)
	if ok {
		return itemer
	}
	return nil
}

func (s *Scene) FindNpcerById(nid int) Npcer {
	npcer, ok := s.sceneObjects[nid].(Npcer)
	if ok {
		return npcer
	}
	return nil
}

func (s *Scene) AllChar() []*Char {
	chars := make([]*Char, 0)
	for _, sb := range s.sceneObjects {
		char, ok := sb.(*Char)
		if ok {
			chars = append(chars, char)
		}
	}
	return chars
}

func (s *Scene) AllBioer() []Bioer {
	bioers := make([]Bioer, 0)
	for _, sb := range s.sceneObjects {
		bioer, ok := sb.(Bioer)
		if ok {
			bioers = append(bioers, bioer)
		}
	}
	return bioers
}

func (s *Scene) AddBody(body *chipmunk.Body) {
	s.cpSpace.AddBody(body)
}

func (s *Scene) RemoveBody(body *chipmunk.Body) {
	s.cpSpace.RemoveBody(body)
}

func (s *Scene) Add(sb SceneObjecter) {
	if sb.Scene() == s {
		return
	}
	sb.SetScene(s)
	sb.SetId(s.idCounter)
	sb.SetInSceneDuration(time.Duration(0))
	s.sceneObjects[s.idCounter] = sb
	s.idCounter = s.idCounter + 1
	s.cpSpace.AddBody(sb.Body())
	sb.OnBeAddedToScene(s)
}

func (s *Scene) Remove(sb SceneObjecter) {
	if sb.Scene() != s {
		return
	}
	delete(s.sceneObjects, sb.Id())
	sb.SetLastId(sb.Id())
	sb.SetLastSceneName(s.name)
	sb.SetScene(nil)
	sb.SetInSceneDuration(time.Duration(0))
	sb.SetId(-1)
	oldBody := sb.Body()
	sb.SetBody(sb.Body().Clone())
	s.cpSpace.RemoveBody(oldBody)
	sb.OnBeRemovedToScene(s)
}
