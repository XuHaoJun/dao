package dao

func NewNpcByBaseId(w *World, id int) *Npc {
	npc := NewNpc(w)
	switch id {
	case 1:
		npc.name = "傳送師"
		npc.baseId = 1
		npc.bodyViewId = 5000
		npcOpt1 := &NpcOption{
			name: "傳送",
			onSelect: func(b Bioer) {
				// TODO
				// trans bio to other scene
				// or change npc talk box
			},
		}
		npc.talk = &NpcTalk{
			title:   npc.name,
			content: "blabla...傳送到野外地圖",
			options: []*NpcOption{
				npcOpt1,
			},
		}
	}
	return npc
}
