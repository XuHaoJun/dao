package dao

type QuestItem struct {
	baseId       int
	currentCount int
	targetCount  int
}

type QuestItemClient struct {
	BaseId       int `json:"baseId"`
	CurrentCount int `json:"currentCount"`
	TargetCount  int `json:"targetCount"`
}

type QuestItemDumpDB struct {
	BaseId       int `bson:"baseId"`
	CurrentCount int `bson:"currentCount"`
	TargetCount  int `bson:"targetCount"`
}

func (qItem *QuestItem) QuestItemClient() *QuestItemClient {
	return &QuestItemClient{
		qItem.baseId,
		qItem.currentCount,
		qItem.targetCount,
	}
}

func (qItem *QuestItem) DumpDB() *QuestItemDumpDB {
	return &QuestItemDumpDB{
		qItem.baseId,
		qItem.currentCount,
		qItem.targetCount,
	}
}

func (qItemDump *QuestItemDumpDB) Load() *QuestItem {
	return &QuestItem{
		qItemDump.BaseId,
		qItemDump.CurrentCount,
		qItemDump.TargetCount,
	}
}

type QuestMob struct {
	baseId       int
	currentCount int
	targetCount  int
}

func (qMob *QuestMob) QuestMobClient() *QuestMobClient {
	return &QuestMobClient{
		qMob.baseId,
		qMob.currentCount,
		qMob.targetCount,
	}
}

func (qMob *QuestMob) DumpDB() *QuestMobDumpDB {
	return &QuestMobDumpDB{
		qMob.baseId,
		qMob.currentCount,
		qMob.targetCount,
	}
}

type QuestMobDumpDB struct {
	BaseId       int `bson:"baseId"`
	CurrentCount int `bson:"currentCount"`
	TargetCount  int `bson:"targetCount"`
}

func (qMobDump *QuestMobDumpDB) Load() *QuestMob {
	return &QuestMob{
		qMobDump.BaseId,
		qMobDump.CurrentCount,
		qMobDump.TargetCount,
	}
}

type QuestMobClient struct {
	BaseId       int `json:"baseId"`
	CurrentCount int `json:"currentCount"`
	TargetCount  int `json:"targetCount"`
}

type QuestFlag struct {
	name string
	flag bool
}

type QuestFlagDumpDB struct {
	Name string `bson:"name"`
	Flag bool   `bson:"flag"`
}

func (qFlag *QuestFlag) DumpDB() *QuestFlagDumpDB {
	return &QuestFlagDumpDB{
		qFlag.name,
		qFlag.flag,
	}
}

func (qFlagDump *QuestFlagDumpDB) Load() *QuestFlag {
	return &QuestFlag{
		qFlagDump.Name,
		qFlagDump.Flag,
	}
}

type QuestPreRequest struct {
	name string
	flag bool
}

type QuestPreRequestDumpDB struct {
	Name string `bson:"name"`
	Flag bool   `bson:"flag"`
}

func (qPreRequest *QuestPreRequest) DumpDB() *QuestPreRequestDumpDB {
	return &QuestPreRequestDumpDB{
		qPreRequest.name,
		qPreRequest.flag,
	}
}

func (qPreReqDump *QuestPreRequestDumpDB) Load() *QuestPreRequest {
	return &QuestPreRequest{
		qPreReqDump.Name,
		qPreReqDump.Flag,
	}
}

type Quest struct {
	baseId      int
	preRequest  []*QuestPreRequest
	targetItems []*QuestItem
	targetMobs  []*QuestMob
	targetFlags []*QuestFlag
	rewards     []*QuestReward
}

type QuestRewardItem struct {
	baseId int
	count  int
}

type QuestReward struct {
	zeny  int
	items []*QuestRewardItem
}

type QuestClient struct {
	BaseId      int                `json:"baseId"`
	TargetItems []*QuestItemClient `json:"targetItems"`
	TargetMobs  []*QuestMobClient  `json:"targetMobs"`
}

type QuestDumpDB struct {
	BaseId      int                      `bson:"baseId"`
	PreRequest  []*QuestPreRequestDumpDB `bson:"preRequest"`
	TargetItems []*QuestItemDumpDB       `bson:"targetItems"`
	TargetMobs  []*QuestMobDumpDB        `bson:"targetMobs"`
	TargetFlags []*QuestFlagDumpDB       `bson:"targetFlags"`
}

func (q *Quest) DumpDB() *QuestDumpDB {
	qDump := &QuestDumpDB{
		q.baseId,
		make([]*QuestPreRequestDumpDB, len(q.preRequest)),
		make([]*QuestItemDumpDB, len(q.targetItems)),
		make([]*QuestMobDumpDB, len(q.targetMobs)),
		make([]*QuestFlagDumpDB, len(q.targetFlags)),
	}
	for i, preReq := range q.preRequest {
		qDump.PreRequest[i] = preReq.DumpDB()
	}
	for i, targetItem := range q.targetItems {
		qDump.TargetItems[i] = targetItem.DumpDB()
	}
	for i, targetMob := range q.targetMobs {
		qDump.TargetMobs[i] = targetMob.DumpDB()
	}
	for i, targetFlag := range q.targetFlags {
		qDump.TargetFlags[i] = targetFlag.DumpDB()
	}
	return qDump
}

func (qDump *QuestDumpDB) Load() *Quest {
	q := &Quest{}
	q.baseId = qDump.BaseId
	q.targetItems = make([]*QuestItem, len(qDump.TargetItems))
	for i, qItemDump := range qDump.TargetItems {
		q.targetItems[i] = qItemDump.Load()
	}
	q.targetMobs = make([]*QuestMob, len(qDump.TargetMobs))
	for i, qMobDump := range qDump.TargetMobs {
		q.targetMobs[i] = qMobDump.Load()
	}
	q.targetFlags = make([]*QuestFlag, len(qDump.TargetFlags))
	for i, qFlagDump := range qDump.TargetFlags {
		q.targetFlags[i] = qFlagDump.Load()
	}
	q.preRequest = make([]*QuestPreRequest, len(qDump.PreRequest))
	for i, qReqDump := range qDump.PreRequest {
		q.preRequest[i] = qReqDump.Load()
	}
	return q
}

func NewQuest() *Quest {
	return &Quest{
		-1,
		make([]*QuestPreRequest, 0),
		make([]*QuestItem, 0),
		make([]*QuestMob, 0),
		make([]*QuestFlag, 0),
		make([]*QuestReward, 0),
	}
}

func (q *Quest) AddMob(id int, targetCount int) {
	qMob := &QuestMob{id, 0, targetCount}
	q.targetMobs = append(q.targetMobs, qMob)
}

func (q *Quest) QuestClient() *QuestClient {
	qItems := make([]*QuestItemClient, len(q.targetItems))
	for i, qItem := range q.targetItems {
		qItems[i] = qItem.QuestItemClient()
	}
	qMobs := make([]*QuestMobClient, len(q.targetMobs))
	for i, qMob := range q.targetMobs {
		qMobs[i] = qMob.QuestMobClient()
	}
	return &QuestClient{q.baseId, qItems, qMobs}
}

func (q *Quest) CanTake() bool {
	if len(q.preRequest) == 0 {
		return true
	}
	for _, req := range q.preRequest {
		if !req.flag {
			return false
		}
	}
	return true
}

func (q *Quest) IncTargetMobCount(mid int, count int) bool {
	for _, tMob := range q.targetMobs {
		if tMob.baseId == mid {
			tMob.currentCount += count
			return true
		}
	}
	return false
}

func (q *Quest) IsComplete() bool {
	for _, questItem := range q.targetItems {
		if questItem.currentCount != questItem.targetCount {
			return false
		}
	}
	for _, questMob := range q.targetMobs {
		if questMob.currentCount != questMob.targetCount {
			return false
		}
	}
	for _, questFlag := range q.targetFlags {
		if !questFlag.flag {
			return false
		}
	}
	return true
}
