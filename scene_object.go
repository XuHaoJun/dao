package dao

import (
	"github.com/xuhaojun/chipmunk"
	"time"
)

type SceneObjecter interface {
	Id() int
	SetId(int)
	Scene() *Scene
	LastSceneName() string
	SetLastSceneName(s string)
	SetLastId(id int)
	LastId() int
	SetScene(*Scene)
	Body() *chipmunk.Body
	SetBody(*chipmunk.Body)
	AfterUpdate(delta float32)
	BeforeUpdate(delta float32)
	OnBeAddedToScene(s *Scene)
	OnBeRemovedToScene(s *Scene)
	SetInSceneDuration(t time.Duration)
	IncInSceneDuration(t time.Duration)
	InSceneDuration() time.Duration
}

type SceneObject struct {
	id              int
	scene           *Scene
	body            *chipmunk.Body
	lastSceneName   string
	lastId          int
	inSceneDuration time.Duration
}

func NewSceneObject() *SceneObject {
	return &SceneObject{}
}

func (sb *SceneObject) Id() int {
	return sb.id
}

func (sb *SceneObject) SetInSceneDuration(t time.Duration) {
	sb.inSceneDuration = t
}

func (sb *SceneObject) IncInSceneDuration(t time.Duration) {
	sb.inSceneDuration += t
}

func (sb *SceneObject) InSceneDuration() time.Duration {
	return sb.inSceneDuration
}

func (sb *SceneObject) SetId(id int) {
	sb.id = id
}

func (sb *SceneObject) SetLastId(id int) {
	sb.lastId = id
}

func (sb *SceneObject) LastId() int {
	return sb.lastId
}

func (sb *SceneObject) SetLastSceneName(s string) {
	sb.lastSceneName = s
}

func (sb *SceneObject) LastSceneName() string {
	return sb.lastSceneName
}

func (sb *SceneObject) Scene() *Scene {
	return sb.scene
}

func (sb *SceneObject) SetScene(s *Scene) {
	sb.scene = s
}

func (sb *SceneObject) Body() *chipmunk.Body {
	return sb.body
}

func (sb *SceneObject) SetBody(b *chipmunk.Body) {
	sb.body = b
}

func (sb *SceneObject) AfterUpdate(delta float32) {
}

func (sb *SceneObject) BeforeUpdate(delta float32) {
}

func (sb *SceneObject) OnBeAddedToScene(s *Scene) {
}

func (sb *SceneObject) OnBeRemovedToScene(s *Scene) {
}
