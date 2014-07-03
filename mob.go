package dao

type Mob struct {
	*BattleBioBase
	id int
}

// default it will only be set by scene type object
func (m *Mob) SetId(id int) {
	m.job <- func() {
		m.id = id
	}
}

func (m *Mob) Id() int {
	idC := make(chan int, 1)
	m.job <- func() {
		idC <- m.id
	}
	return <-idC
}
