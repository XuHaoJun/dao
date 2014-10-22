package dao

type Npc struct {
	*Bio
	baseId          int
	talk            *NpcTalk
	OnFirstBeTalked func(b Bioer)
}

type NpcTalk struct {
	title   string
	content string
	options []*NpcOption
}

func (nt *NpcTalk) NpcTalkClient() *NpcTalkClient {
	ntClient := &NpcTalkClient{
		Title:   nt.title,
		Content: nt.content,
		Options: make([]*NpcOptionClient, len(nt.options)),
	}
	for i, npcOpt := range nt.options {
		ntClient.Options[i] = npcOpt.NpcOptionClient()
	}
	return ntClient
}

type NpcTalkClient struct {
	Title   string             `json:"title"`
	Content string             `json:"content"`
	Options []*NpcOptionClient `json:"options"`
}

type NpcOption struct {
	name        string
	onSelect    func(b Bioer)
	nextNpcTalk *NpcTalk
}

func (no *NpcOption) NpcOptionClient() *NpcOptionClient {
	return &NpcOptionClient{
		Name: no.name,
	}
}

type NpcOptionClient struct {
	Name string `json:"name"`
}

type Npcer interface {
	Bioer
	FirstBeTalked(b Bioer) bool
	SelectOption(optIndex int, b Bioer)
	NpcClientBasic() *NpcClientBasic
	NpcTalkClient() *NpcTalkClient
	Bioer() Bioer
}

type TalkingNpcInfo struct {
	target  Npcer
	options []int
}

//type NpcOptionNum int

type NpcClientBasic struct {
	BioClient *BioClientBasic `json:"bioConfig"`
}

func NewNpc(w *World) *Npc {
	npc := &Npc{
		Bio:    NewBio(w),
		baseId: -1,
	}
	npc.bodyViewId = 5000
	npc.body.UserData = npc
	return npc
}

func (n *Npc) Bioer() Bioer {
	return n
}

func (n *Npc) Npcer() Npcer {
	return n
}

func (n *Npc) SceneObjecter() SceneObjecter {
	return n
}

func (n *Npc) PublishClientCall(cc *ClientCall) {
	n.scene.DispatchClientCall(n, cc)
}

func (n *Npc) NpcClientBasic() *NpcClientBasic {
	bClient := n.Bio.BioClientBasic()
	return &NpcClientBasic{
		BioClient: bClient,
	}
}

func (n *Npc) FirstBeTalked(b Bioer) bool {
	// TODO show talk box to char client
	// or do something when talk
	if b.TalkingNpcInfo().target == n.Npcer() {
		return false
	}
	if n.OnFirstBeTalked != nil {
		n.OnFirstBeTalked(b)
	}
	talkingNpcInfo := b.TalkingNpcInfo()
	talkingNpcInfo.target = n.Npcer()
	return true
}

func (n *Npc) NpcTalkClient() *NpcTalkClient {
	return n.talk.NpcTalkClient()
}

func (n *Npc) SelectOption(optIndex int, b Bioer) {
	n.talk.options[optIndex].onSelect(b)
}

func (n *Npc) OnBeRemovedToScene(s *Scene) {
	n.Bio.OnBeRemovedToScene(s)
	bioers := s.AllBioer()
	for _, bioer := range bioers {
		if bioer.TalkingNpcInfo().target == n.Npcer() {
			bioer.SetTalkingNpcInfo(nil)
			// TODO
			// cancel talk box to bioer client
			return
		}
	}
}
