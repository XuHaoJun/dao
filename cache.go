package dao

import (
	"reflect"
)

type Cache struct {
	WorldClientCallMethods map[string]reflect.Value
	UseSelfFuncs           map[int]func(b Bioer)
}

func NewCache() *Cache {
	return &Cache{
		WorldClientCallMethods: make(map[string]reflect.Value),
		UseSelfFuncs:           make(map[int]func(b Bioer)),
	}
}
