package dao

import (
	"github.com/xuhaojun/chipmunk"
)

type SceneObjecter interface {
	Id() int
	SetId(int)
	Scene() *Scene
	LastSceneName() string
	LastId() int
	SetScene(*Scene)
	Body() *chipmunk.Body
	SetBody(*chipmunk.Body)
	AfterUpdate(delta float32)
	BeforeUpdate(delta float32)
	OnBeAddedToScene(s *Scene)
	OnBeRemovedToScene(s *Scene)
}

type SceneObject struct {
	id            int
	scene         *Scene
	body          *chipmunk.Body
	lastSceneName string
	lastId        int
}

func NewSceneObject() *SceneObject {
	return &SceneObject{}
}

func (sb *SceneObject) Id() int {
	return sb.id
}

func (sb *SceneObject) SetId(id int) {
	sb.id = id
}

func (sb *SceneObject) LastId() int {
	return sb.lastId
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
