package dao

type Cache struct {
	UseSelfFuncs map[int]func(b Bioer)
}

func NewCache() *Cache {
	return &Cache{
		UseSelfFuncs: make(map[int]func(b Bioer)),
	}
}
