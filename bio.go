package dao

import (
	"fmt"
	"github.com/xuhaojun/chipmunk"
	"github.com/xuhaojun/chipmunk/vect"
	"math"
)

var (
	BioGroup = chipmunk.Group(1)
	BioLayer = chipmunk.Layer(1)
)

type Bioer interface {
	World() *World
	Name() string
	Id() int
	SetId(int)
	Move(x, y float32)
	ShutDownMove()
	Body() *chipmunk.Body
	SetBody(*chipmunk.Body)
	Scene() *Scene
	SetScene(*Scene)
	BioClient() *BioClient
	BioClientBasic() *BioClientBasic
	BioClientAttributes() *BioClientAttributes
	SceneObjecter() SceneObjecter
	SetPosition(float32, float32)
	CpBody() *chipmunk.Body
	LookAtByBioer(target Bioer)
	// npc
	TalkingNpcInfo() *TalkingNpcInfo
	SetTalkingNpcInfo(*TalkingNpcInfo)
	CancelTalkingNpc()
	ResponseTalkingNpc(optIndex int)
	// teleport
	TeleportBySceneName(name string, x float32, y float32) *Scene
	//
	Str() int
	Spi() int
	Vit() int
	Wis() int
	Hp() int
	Level() int
	//
	IsDied() bool
	TakeDamage(d int, b Bioer)
	Reborn()
	//
	OnKillFunc() func(target Bioer)
	// imple Stringer
	String() string
}

type Bio struct {
	*SceneObject
	world      *World
	name       string
	bodyViewId int
	//
	moveState *MoveState
	//
	clientCallPublisher ClientCallPublisher
	skillUser           Bioer
	beKilleder          Bioer
	//
	// aoi
	viewAOIState *ViewAOIState
	//
	level int
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
	OnKill     func(target Bioer)
	OnBeKilled func(killer Bioer)
	// skills
	fireBallState *FireBallState
	// npc interactive
	talkingNpcInfo *TalkingNpcInfo
}

type BioClient struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	BodyViewId int    `json:"bodyViewId"`
	Level      int    `json:"level"`
	// main attribue
	Str int `json:"str"`
	Vit int `json:"vit"`
	Wis int `json:"wis"`
	Spi int `json:"spi"`
	// sub attribue
	Def   int `json:"def"`
	Mdef  int `json:"mdef"`
	Atk   int `json:"atk"`
	Matk  int `json:"matk"`
	MaxHp int `json:"maxHp"`
	Hp    int `json:"hp"`
	MaxMp int `json:"maxMp"`
	Mp    int `json:"mp"`
	//
	MoveBaseVelocity *CpVectClient `json:"moveBaseVelocity"`
	//
	CpBody *CpBodyClient `json:"cpBody"`
}

type BioClientBasic struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	BodyViewId int    `json:"bodyViewId"`
	Level      int    `json:"level"`
	//
	MaxHp int `json:"maxHp"`
	Hp    int `json:"hp"`
	MaxMp int `json:"maxMp"`
	Mp    int `json:"mp"`
	//
	MoveBaseVelocity *CpVectClient `json:"moveBaseVelocity"`
	//
	CpBody *CpBodyClient `json:"cpBody"`
}

type BioClientAttributes struct {
	// main attribute
	Str int `json:"str"`
	Vit int `json:"vit"`
	Wis int `json:"wis"`
	Spi int `json:"spi"`
	// sub attribute
	Def   int `json:"def"`
	Mdef  int `json:"mdef"`
	Atk   int `json:"atk"`
	Matk  int `json:"matk"`
	MaxHp int `json:"maxHp"`
	Hp    int `json:"hp"`
	MaxMp int `json:"maxMp"`
	Mp    int `json:"mp"`
}

type CpVectClient struct {
	X vect.Float `json:"x"`
	Y vect.Float `json:"y"`
}

type CpBodyClient struct {
	Mass     vect.Float    `json:"mass"`
	Angle    vect.Float    `json:"angle"`
	Shapes   []interface{} `json:"shapes"`
	Position *CpVectClient `json:"position"`
	Velocity *CpVectClient `json:"velocity"`
}

type ViewAOIState struct {
	running bool
	//
	stepDone bool
	//
	body *chipmunk.Body
	//
	radius float32
	//
	inAreaSceneObjecters map[SceneObjecter]struct{}
	// callbacks
	OnSceneObjectEnter func(sb SceneObjecter)
	OnSceneObjectLeave func(sb SceneObjecter)
}

// type InAreaSceneObjecters map[SceneObjecter]struct{}

// func (sbs InAreaSceneObjecters) FindBioer(b Bioer) Bioer {
// 	_, ok := sbs[b.SceneObjecter()]
// 	if ok {
// 		return b
// 	}
// 	return nil
// }

// func (sbs InAreaSceneObjecters) FindSceneObjecter(b SceneObjecter) SceneObjecter {
// 	_, ok := sbs[b]
// 	if ok {
// 		return b
// 	}
// 	return nil
// }

func NewViewAOIState() *ViewAOIState {
	return &ViewAOIState{
		running: false,
	}
}

type MoveState struct {
	running          bool
	beforeMoveFunc   func(delta float32) bool
	targetPos        vect.Vect
	lastTargetPos    vect.Vect
	baseVelocity     vect.Vect
	lastBaseVelocity vect.Vect
}

type MoveStateClient struct {
	Running      bool          `json:"running"`
	TargetPos    *CpVectClient `json:"targetPos"`
	BaseVelocity *CpVectClient `json:"baseVelocity,omitempty"`
}

func (m *MoveState) MoveStateClient() *MoveStateClient {
	return &MoveStateClient{
		Running: m.running,
		TargetPos: &CpVectClient{
			m.targetPos.X,
			m.targetPos.Y,
		},
		BaseVelocity: &CpVectClient{
			m.baseVelocity.X,
			m.baseVelocity.Y,
		},
	}
}

func NewBioViewAOIState(r float32, bio *Bio) *ViewAOIState {
	viewAOIState := &ViewAOIState{
		running:              true,
		stepDone:             false,
		radius:               r,
		inAreaSceneObjecters: make(map[SceneObjecter]struct{}),
	}
	viewAOIState.OnSceneObjectEnter = bio.OnSceneObjectEnterViewAOIFunc()
	viewAOIState.OnSceneObjectLeave = bio.OnSceneObjectLeaveViewAOIFunc()
	viewAOIBody := chipmunk.NewBody(1, 1)
	viewAOIBody.CallbackHandler = viewAOIState
	viewAOIShape := chipmunk.NewCircle(vect.Vector_Zero, viewAOIState.radius)
	viewAOIShape.IsSensor = true
	viewAOIBody.AddShape(viewAOIShape)
	viewAOIState.body = viewAOIBody
	return viewAOIState
}

func NewBio(w *World) *Bio {
	body := chipmunk.NewBody(1, 1)
	circle := chipmunk.NewCircle(vect.Vector_Zero, float32(32.0))
	circle.Group = BioGroup
	circle.Layer = BioLayer
	circle.SetFriction(0)
	circle.SetElasticity(0)
	body.SetPosition(vect.Vector_Zero)
	body.SetVelocity(0, 0)
	body.SetMoment(chipmunk.Inf)
	body.IgnoreGravity = true
	body.AddShape(circle)
	bio := &Bio{
		SceneObject: &SceneObject{
			body: body,
		},
		world: w,
		name:  "",
		moveState: &MoveState{
			running:      false,
			baseVelocity: vect.Vect{X: 90, Y: 90},
		},
		str: 1,
		vit: 1,
		wis: 1,
		spi: 1,
		//
		talkingNpcInfo: &TalkingNpcInfo{
			target:  nil,
			options: make([]int, 0),
		},
	}
	body.UserData = bio
	bio.viewAOIState = NewBioViewAOIState(1000, bio)
	bio.clientCallPublisher = bio.ClientCallPublisher()
	bio.skillUser = bio.Bioer()
	bio.beKilleder = bio.Bioer()
	bio.fireBallState = NewFireBallState(bio.skillUser)
	return bio
}

func (b *Bio) TalkingNpcInfo() *TalkingNpcInfo {
	return b.talkingNpcInfo
}

func (b *Bio) String() string {
	s := fmt.Sprintf("Name: %s\n", b.name)
	s += fmt.Sprintf("Level: %d\n", b.level)
	if b.scene != nil {
		s += "Position: " + b.scene.name + " "
		s += b.body.Position().String()
	}
	return s
}

func (b *Bio) SetTalkingNpcInfo(tNpc *TalkingNpcInfo) {
	b.talkingNpcInfo = tNpc
}

func (b *Bio) CancelTalkingNpc() {
	if b.talkingNpcInfo.target != nil {
		b.talkingNpcInfo = &TalkingNpcInfo{
			target:  nil,
			options: make([]int, 0),
		}
	}
}

func (b *Bio) ResponseTalkingNpc(optIndex int) {
	if optIndex < 0 || b.talkingNpcInfo.target == nil {
		return
	}
	npc := b.talkingNpcInfo.target
	npc.SelectOption(optIndex, b.Bioer())
}

func (b *Bio) SetPosition(x float32, y float32) {
	b.body.SetPosition(
		vect.Vect{X: vect.Float(x),
			Y: vect.Float(y)})
	if b.viewAOIState.running && b.viewAOIState.body != nil {
		b.viewAOIState.body.SetPosition(
			vect.Vect{X: vect.Float(x),
				Y: vect.Float(y)})
	}
}

func (b *Bio) CpBody() *chipmunk.Body {
	return b.SceneObject.body
}

func (b *Bio) LookAtByBioer(target Bioer) {
	cpBody := target.CpBody()
	b.SceneObject.body.LookAt(cpBody.Position())
}

func (b *Bio) OnBeAddedToScene(s *Scene) {
	b.viewAOIState.body.SetPosition(b.body.Position())
	s.AddBody(b.viewAOIState.body)
}

func (b *Bio) OnBeRemovedToScene(s *Scene) {
	s.RemoveBody(b.viewAOIState.body)
	old := b.viewAOIState
	b.viewAOIState = NewBioViewAOIState(old.radius, b)
	if old.OnSceneObjectEnter != nil {
		b.viewAOIState.OnSceneObjectEnter = old.OnSceneObjectEnter
	}
	if old.OnSceneObjectLeave != nil {
		b.viewAOIState.OnSceneObjectLeave = old.OnSceneObjectLeave
	}
}

func (b *Bio) OnSceneObjectEnterViewAOIFunc() func(sb SceneObjecter) {
	return func(enter SceneObjecter) {
		// new bio enter may be attack it if this is mob
		// new bio enter may be Update to client screen it if this is mob
	}
}

func (b *Bio) OnSceneObjectLeaveViewAOIFunc() func(sb SceneObjecter) {
	return func(leaver SceneObjecter) {
	}
}

func (b *Bio) SceneObjecter() SceneObjecter {
	return b
}

func (b *Bio) Bioer() Bioer {
	return b
}

func (b *Bio) Move(x, y float32) {
	if b.IsDied() {
		return
	}
	b.moveState.running = true
	b.moveState.targetPos = vect.Vect{
		X: vect.Float(x),
		Y: vect.Float(y),
	}
}

func (b *Bio) ShutDownMove() {
	b.moveState.running = false
	b.body.SetForce(0, 0)
	b.body.SetVelocity(0, 0)
	b.PublishMoveState()
	// logger := b.scene.world.logger
	// logger.Println("position", b.body.Position())
	// logger.Println("velocity", b.body.Velocity())
}

func (b *Bio) PublishMoveState() {
	clientCall := &ClientCall{
		Receiver: "bio",
		Method:   "handleMoveStateChange",
		Params: []interface{}{
			b.id,
			b.moveState.MoveStateClient(),
		},
	}
	b.clientCallPublisher.PublishClientCall(clientCall)
}

func (b *Bio) MoveUpdate(delta float32) {
	if b.moveState.running == false {
		return
	}
	if !vect.Equals(b.moveState.targetPos, b.moveState.lastTargetPos) ||
		!vect.Equals(b.moveState.lastBaseVelocity, b.moveState.baseVelocity) {
		b.PublishMoveState()
	}
	if vect.Equals(b.body.Position(), b.moveState.targetPos) {
		b.ShutDownMove()
		return
	}
	if b.moveState.beforeMoveFunc != nil {
		keepMove := b.moveState.beforeMoveFunc(delta)
		if keepMove == false {
			return
		}
	}
	moveVelocity := b.moveState.baseVelocity
	cpBodyPos := b.body.Position()
	moveVect := vect.Vect{
		X: b.moveState.targetPos.X - cpBodyPos.X,
		Y: b.moveState.targetPos.Y - cpBodyPos.Y,
	}
	if math.Abs(float64(delta*float32(moveVelocity.X))) >=
		math.Abs(float64(moveVect.X)) {
		cpBodyPos = b.body.Position()
		reachX := vect.Vect{
			X: b.moveState.targetPos.X,
			Y: cpBodyPos.Y,
		}
		b.body.SetPosition(reachX)
		moveVect.X = 0
	}
	if math.Abs(float64(delta*float32(moveVelocity.Y))) >=
		math.Abs(float64(moveVect.Y)) {
		cpBodyPos = b.body.Position()
		reachY := vect.Vect{
			X: cpBodyPos.X,
			Y: b.moveState.targetPos.Y,
		}
		b.body.SetPosition(reachY)
		moveVect.Y = 0
	}
	if moveVect.X < 0 {
		moveVelocity.X *= -1
	} else if moveVect.X == 0 {
		moveVelocity.X = 0
	}
	if moveVect.Y < 0 {
		moveVelocity.Y *= -1
	} else if moveVect.Y == 0 {
		moveVelocity.Y = 0
	}
	cpBody := b.body
	vel := cpBody.Velocity()
	m := cpBody.Mass()
	t := delta
	desiredVel := moveVelocity
	velChange := vect.Vect{
		X: desiredVel.X - vel.X,
		Y: desiredVel.Y - vel.Y,
	}
	force := vect.Vect{
		X: m * velChange.X / vect.Float(t),
		Y: m * velChange.Y / vect.Float(t),
	}
	cpBody.SetForce(float32(force.X), float32(force.Y))
	cpBody.LookAt(b.moveState.targetPos)
	b.moveState.lastTargetPos = b.moveState.targetPos
	b.moveState.lastBaseVelocity = b.moveState.baseVelocity
	// logger := b.scene.world.logger
	// logger.Println("moveVect", moveVect)
	// logger.Println("position", b.body.Position())
	// logger.Println("velocity", b.body.Velocity())
	// logger.Println("degree", b.body.Angle()*(180/math.Pi))
	// logger.Println("delta", delta)
	// logger.Println("force", force)
}

func (b *Bio) BeforeUpdate(delta float32) {
}

func (b *Bio) AfterUpdate(delta float32) {
	b.MoveUpdate(delta)
	b.ViewAOIUpdate(delta)
	b.FireBallUpdate(delta)
}

func (b *Bio) Name() string {
	return b.name
}

func (b *Bio) World() *World {
	return b.world
}

func (b *Bio) ClientCallPublisher() ClientCallPublisher {
	return b
}

func (b *Bio) PublishClientCall(c *ClientCall) {
	if b.scene == nil {
		return
	}
	b.scene.DispatchClientCall(b, c)
}

func (b *Bio) Level() int {
	return b.level
}

func (b *Bio) Hp() int {
	return b.hp
}

func (b *Bio) Str() int {
	return b.str
}

func (b *Bio) Vit() int {
	return b.vit
}

func (b *Bio) Wis() int {
	return b.wis
}

func (b *Bio) Spi() int {
	return b.spi
}

func (b *Bio) CalcAttributes() {
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

func (b *Bio) IncHp(n int) {
	if b.hp <= 0 || n < 0 {
		return
	}
	tmpHp := b.hp
	tmpHp += n
	if tmpHp >= b.maxHp {
		return
	} else {
		b.hp = tmpHp
	}
}

func (b *Bio) DecHp(n int, killer Bioer) bool {
	if b.hp <= 0 || n < 0 {
		return false
	}
	tmpHp := b.hp
	tmpHp -= n
	if tmpHp <= 0 {
		b.hp = 0
		if b.OnBeKilled != nil {
			b.OnBeKilled(killer)
		}
		return true
	} else {
		b.hp = tmpHp
	}
	return false
}

func (b *Bio) DecMp(n int) {
	if b.mp <= 0 || b.IsDied() {
		return
	}
	tmpMp := b.mp
	tmpMp -= n
	if tmpMp <= 0 {
		b.mp = 0
	} else {
		b.mp = tmpMp
	}
}

func (b *Bio) IsDied() bool {
	return b.hp <= 0
}

func (b *Bio) BioClient() *BioClient {
	return &BioClient{
		Id:         b.id,
		Name:       b.name,
		Level:      b.level,
		BodyViewId: b.bodyViewId,
		//
		Str: b.str,
		Vit: b.vit,
		Wis: b.wis,
		Spi: b.spi,
		//
		Def:   b.def,
		Mdef:  b.mdef,
		Atk:   b.atk,
		Matk:  b.matk,
		MaxHp: b.maxHp,
		Hp:    b.hp,
		MaxMp: b.maxMp,
		Mp:    b.mp,
		//
		MoveBaseVelocity: &CpVectClient{
			b.moveState.baseVelocity.X,
			b.moveState.baseVelocity.Y,
		},
		//
		CpBody: ToCpBodyClient(b.body),
	}
}

func (b *Bio) BioClientBasic() *BioClientBasic {
	return &BioClientBasic{
		Id:         b.id,
		Name:       b.name,
		Level:      b.level,
		BodyViewId: b.bodyViewId,
		//
		Hp:    b.hp,
		MaxHp: b.maxHp,
		MaxMp: b.maxMp,
		Mp:    b.mp,
		//
		MoveBaseVelocity: &CpVectClient{
			b.moveState.baseVelocity.X,
			b.moveState.baseVelocity.Y,
		},
		//
		CpBody: ToCpBodyClient(b.body),
	}
}

func (b *Bio) BioClientAttributes() *BioClientAttributes {
	return &BioClientAttributes{
		Str:   b.str,
		Vit:   b.vit,
		Wis:   b.wis,
		Spi:   b.spi,
		Def:   b.def,
		Mdef:  b.mdef,
		Atk:   b.atk,
		Matk:  b.matk,
		MaxHp: b.maxHp,
		Hp:    b.hp,
		MaxMp: b.maxMp,
		Mp:    b.mp,
	}
}

type ChatMessageClient struct {
	ChatType string `json:"chatType"`
	Talker   string `json:"talker"`
	Content  string `json:"content"`
}

func (b *Bio) UseFireBall() {
	if b.IsDied() || b.scene == nil {
		return
	}
	b.fireBallState.Fire()
}

func (b *Bio) Reborn() {
	return
}

func (b *Bio) TeleportBySceneName(name string, x float32, y float32) (targetScene *Scene) {
	curScene := b.scene
	targetScene = b.world.FindSceneByName(name)
	if curScene == nil ||
		targetScene == nil {
		return
	}
	if curScene == targetScene {
		b.SetPosition(x, y)
		return curScene
	}
	b.lastSceneName = curScene.name
	b.lastId = b.id
	curScene.Remove(b.SceneObjecter())
	b.SetPosition(x, y)
	targetScene.Add(b.SceneObjecter())
	return
}

func (b *Bio) TalkScene(content string) {
	clientCall := &ClientCall{
		Receiver: "char",
		Method:   "handleChatMessage",
		Params: []interface{}{
			&ChatMessageClient{
				"Scene",
				b.name,
				content,
			},
		},
	}
	b.scene.DispatchClientCall(b, clientCall)
}

func (b *Bio) TakeDamage(d int, attacker Bioer) {
	// server
	if b.IsDied() {
		return
	}
	b.hp -= d
	if b.hp < 0 {
		b.hp = 0
	}
	//client update
	clientCall := &ClientCall{
		Receiver: "bio",
		Method:   "handleUpdateBioConfig",
		Params: []interface{}{
			b.id,
			map[string]int{
				"hp": b.hp,
			},
		},
	}
	b.clientCallPublisher.PublishClientCall(clientCall)
	if b.hp == 0 {
		b.lastId = b.id
		b.lastSceneName = b.scene.name
		scene := b.scene
		scene.Remove(b)
		if b.OnBeKilled != nil {
			b.OnBeKilled(attacker)
		}
		onkill := attacker.OnKillFunc()
		if onkill != nil {
			onkill(b.beKilleder)
		}
	}
}

func (b *Bio) OnKillFunc() func(target Bioer) {
	return b.OnKill
}

func (b *Bio) ItemQuickHeal(n int, effectId int) bool {
	if b.IsDied() {
		return false
	}
	b.hp += n
	if b.hp > b.maxHp {
		b.hp = b.maxHp
	}
	b.world.logger.Println("ItemQuickHeal")
	clientCall1 := &ClientCall{
		Receiver: "bio",
		Method:   "handleItemQuickHeal",
		Params: []interface{}{
			b.id,
			n,
			effectId,
		},
	}
	clientCall2 := &ClientCall{
		Receiver: "bio",
		Method:   "handleUpdateBioConfig",
		Params: []interface{}{
			b.id,
			map[string]int{
				"hp": b.hp,
			},
		},
	}
	b.clientCallPublisher.PublishClientCall(clientCall1)
	b.clientCallPublisher.PublishClientCall(clientCall2)
	return true
}

func (b *Bio) ViewAOIUpdate(delta float32) {
	if b.viewAOIState == nil {
		return
	}
	b.viewAOIState.body.SetPosition(b.body.Position())
}

func (b *Bio) FireBallUpdate(delta float32) {
	if b.fireBallState == nil {
		return
	}
	b.fireBallState.Update(delta)
}

func (v *ViewAOIState) CollisionEnter(arbiter *chipmunk.Arbiter) bool {
	sb, ok := arbiter.BodyA.UserData.(SceneObjecter)
	if ok {
		v.inAreaSceneObjecters[sb] = struct{}{}
		if v.OnSceneObjectEnter != nil {
			v.OnSceneObjectEnter(sb)
		}
	}
	sb, ok = arbiter.BodyB.UserData.(SceneObjecter)
	if ok {
		v.inAreaSceneObjecters[sb] = struct{}{}
		if v.OnSceneObjectEnter != nil {
			v.OnSceneObjectEnter(sb)
		}
	}
	return false
}

func (v *ViewAOIState) CollisionExit(arbiter *chipmunk.Arbiter) {
	sb, ok := arbiter.BodyA.UserData.(SceneObjecter)
	if ok {
		delete(v.inAreaSceneObjecters, sb)
		if v.OnSceneObjectLeave != nil {
			v.OnSceneObjectLeave(sb)
		}
	}
	sb, ok = arbiter.BodyB.UserData.(SceneObjecter)
	if ok {
		delete(v.inAreaSceneObjecters, sb)
		if v.OnSceneObjectLeave != nil {
			v.OnSceneObjectLeave(sb)
		}
	}
}

func (v *ViewAOIState) CollisionPreSolve(arbiter *chipmunk.Arbiter) bool {
	return false
}

func (v *ViewAOIState) CollisionPostSolve(arbiter *chipmunk.Arbiter) {}
