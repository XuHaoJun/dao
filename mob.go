package dao

type Mob struct {
	*BioBase
	id   int
	name string
	str  int
	vit  int
	wis  int
	spi  int
}
