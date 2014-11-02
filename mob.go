package dao

type Mober interface {
	Bioer
}

type Mob struct {
	*Bio
	baseId int
}

func NewMob(w *World) *Mob {
	mob := &Mob{
		Bio:    NewBio(w),
		baseId: -1,
	}
	mob.bodyViewId = 10001
	mob.body.UserData = mob
	return mob
}

func (m *Mob) Bioer() Bioer {
	return m
}

func (m *Mob) Mober() Mober {
	return m
}

func (m *Mob) SceneObjecter() SceneObjecter {
	return m
}
