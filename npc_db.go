package dao

func NewNpcByBaseId(w *World, id int) Npcer {
	npc := NewNpc(w)
	npc.baseId = id
	switch id {
	case 1:
		npc.name = "傳送師"
		npc.bodyViewId = 5000
		npcOpt1 := &NpcOption{
			name: "傳送",
			onSelect: func(curNpc Npcer, nextNpcTalk *NpcTalk, b Bioer) {
				if nextNpcTalk == nil {
					switch c := b.(type) {
					case Charer:
						c.TeleportBySceneName("daoField01", 0, 0)
						c.CancelTalkingNpc()
					default:
						b.CancelTalkingNpc()
					}
					return
				}
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
	case 2:
		npc.name = "Jack"
		npc.bodyViewId = 5000
		npc.shop = &Shop{"Jack's Shop",
			[]int{1, 2, 5001, 5002, 10001, 5003},
			npc.Bioer(),
			npc.world}
		npcOpt0 := &NpcOption{
			name: "Hello",
			nextNpcTalk: &NpcTalk{
				title:   npc.name,
				content: "hello hello hello hello.............",
			},
			onSelect: func(curNpc Npcer, nextNpcTalk *NpcTalk, b Bioer) {
				if nextNpcTalk == nil {
					switch c := b.(type) {
					case Charer:
						c.CancelTalkingNpc()
					default:
						b.CancelTalkingNpc()
					}
					return
				}
				tNpcInfo := b.TalkingNpcInfo()
				tNpcInfo.options = append(tNpcInfo.options, 1)
				c, isCharer := b.(Charer)
				if isCharer {
					c.SendNpcTalkBox(nextNpcTalk)
					c.GetItemByBaseId(10001)
					c.GetItemByBaseId(5001)
				}
			},
		}
		npcOpt1 := &NpcOption{
			name: "HaHa",
			nextNpcTalk: &NpcTalk{
				title:   npc.name,
				content: "My name is Jack!",
			},
			onSelect: func(curNpc Npcer, nextNpcTalk *NpcTalk, b Bioer) {
				if nextNpcTalk == nil {
					switch c := b.(type) {
					case Charer:
						c.CancelTalkingNpc()
					default:
						b.CancelTalkingNpc()
					}
					return
				}
				tNpcInfo := b.TalkingNpcInfo()
				tNpcInfo.options = append(tNpcInfo.options, 1)
				c, isCharer := b.(Charer)
				if isCharer {
					clientCall := &ClientCall{
						Receiver: "char",
						Method:   "handleNpcTalkBox",
						Params:   []interface{}{nextNpcTalk.NpcTalkClient()},
					}
					c.SendClientCall(clientCall)
				}
			},
		}
		npcOpt2 := &NpcOption{
			name: "Shop!",
			onSelect: func(curNpc Npcer, nextNpcTalk *NpcTalk, b Bioer) {
				if nextNpcTalk == nil {
					switch c := b.(type) {
					case Charer:
						c.OpenShop(curNpc.Shoper())
						c.CancelTalkingNpc()
					default:
						b.CancelTalkingNpc()
					}
					return
				}
			},
		}
		npc.talk = &NpcTalk{
			title:   npc.name,
			content: "",
			options: []*NpcOption{
				npcOpt0,
				npcOpt1,
				npcOpt2,
			},
		}
		npc.OnFirstBeTalked = func(curNpc Npcer, b Bioer) {
			nTalk := curNpc.NpcTalk()
			nTalk.content = b.Name() + " Hello!"
		}
	}
	return npc
}
