package dao

// TODO
// may be try to save it to
// db or json or yaml for custom it.

func NewEquipmentByBaseId(id int) *Equipment {
	eq := NewEquipment()
	switch id {
	case 1:
		eq.baseId = 1
		eq.iconViewId = 1
		eq.name = "Sword001"
		eq.ageisName = "Sword001"
		eq.bonusInfo = &EquipmentBonusInfo{
			atk: 10,
		}
		eq.etype = Sword
	}
	return eq
}
