package dao

type CommonQuestItem struct {
	currentCount int
	targetCount  int
}

type CommonQuestMob struct {
	currentCount int
	targetCount  int
}

type CommonQuest struct {
	baseId      int
	targetItems map[int]*CommonQuestItem
	targetMobs  map[int]*CommonQuestMob
}

func NewCommonQuest() *CommonQuest {
	return &CommonQuest{
		-1,
		make(map[int]*CommonQuestItem),
		make(map[int]*CommonQuestMob),
	}
}

func (cquest *CommonQuest) IsComplete() bool {
	completeItem := true
	for _, questItem := range cquest.targetItems {
		if questItem.currentCount != questItem.targetCount {
			completeItem = false
			break
		}
	}
	completeMob := true
	for _, questMob := range cquest.targetMobs {
		if questMob.currentCount != questMob.targetCount {
			completeMob = false
			break
		}
	}
	return completeItem && completeMob
}
