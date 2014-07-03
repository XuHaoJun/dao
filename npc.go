package dao

type Npc struct {
	*BioBase
	id int
}

// default it will only be set by scene type object
func (n *Npc) SetId(id int) {
	n.job <- func() {
		n.id = id
	}
}

func (n *Npc) Id() int {
	idC := make(chan int, 1)
	n.job <- func() {
		idC <- n.id
	}
	return <-idC
}
