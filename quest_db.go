package dao

func NewQuestByBaseId(id int) *Quest {
	switch id {
	case 1:
		return newQuestByBaseId1()
	}
	return nil

}
func newQuestByBaseId1() *Quest {
	q := NewQuest()
	q.baseId = 1
	q.AddMob(1, 10)
	q.rewards = append(q.rewards, &QuestReward{1000, nil})
	return q
}
