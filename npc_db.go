package dao

func NewNpcByBaseId(w *World, id int) Npcer {
	npc := NewNpc(w)
	npc.baseId = id
	switch id {
	case 1:
		npc.name = "傳送師"
		npc.bodyViewId = 5000
		npcOpt0 := &NpcOption{
			key:  0,
			name: "傳送",
			onSelect: func(event NpcOptionSelectEvent) {
				b := event.TargetBio
				nextNpcTalk := event.NextNpcTalk
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
				npcOpt0,
			},
		}
	case 2:
		npc.name = "Jack"
		npc.bodyViewId = 5000
		npc.shop = &Shop{"Jack's Shop",
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
				5001, 5002, 5003, 5004,
				10001, 10002},
			npc.Bioer(),
			npc.world}
		npcOpt0 := &NpcOption{
			key:  0,
			name: "Hello",
			nextNpcTalk: &NpcTalk{
				title:   npc.name,
				content: "hello hello hello hello.............",
			},
			onSelect: func(event NpcOptionSelectEvent) {
				b := event.TargetBio
				nextNpcTalk := event.NextNpcTalk
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
				tNpcInfo.options = append(tNpcInfo.options, 0)
				c, isCharer := b.(Charer)
				if isCharer {
					c.SendNpcTalkBox(nextNpcTalk)
					c.GetItemByBaseId(10001)
					c.GetItemByBaseId(5001)
				}
			},
		}
		npcOpt1 := &NpcOption{
			key:  1,
			name: "HaHa",
			nextNpcTalk: &NpcTalk{
				title:   npc.name,
				content: "My name is Jack!",
				options: []*NpcOption{
					npcOpt0,
				},
			},
			onSelect: func(event NpcOptionSelectEvent) {
				b := event.TargetBio
				nextNpcTalk := event.NextNpcTalk
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
			key:  2,
			name: "First Quest!",
			onSelect: func(event NpcOptionSelectEvent) {
				b := event.TargetBio
				curNpc := event.CurrentNpc
				switch c := b.(type) {
				case Charer:
					questBaseId := 1
					quest, found := c.FindQuest(questBaseId)
					if !found {
						c.TakeQuest(NewQuestByBaseId(questBaseId))
					} else {
						if quest.IsComplete() {
							c.ClearQuest(questBaseId)
							c.SendChatMessage("System", curNpc.Name(), "You completed quest!")
						} else {
							c.SendChatMessage("System", curNpc.Name(), "You not complete quest!")
						}
					}
					c.CancelTalkingNpc()
				default:
					b.CancelTalkingNpc()
				}
			},
		}
		npcOpt3 := &NpcOption{
			key:  3,
			name: "Shop!",
			onSelect: func(event NpcOptionSelectEvent) {
				b := event.TargetBio
				nextNpcTalk := event.NextNpcTalk
				curNpc := event.CurrentNpc
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
				npcOpt3,
			},
		}
		npc.OnFirstBeTalked = func(curNpc Npcer, b Bioer) {
			nTalk := curNpc.NpcTalk()
			nTalk.content = b.Name() + " Hello!"
		}
	}
	return npc
}
