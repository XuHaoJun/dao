package dao

type Npc struct {
	*BioBase
}

func NewNpc() *Npc {
	npc := &Npc{
		BioBase: NewBioBase(),
	}
	npc.BioBase.enableViewAOI = false
	return npc
}
