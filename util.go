package dao

import (
	"math/rand"
	"time"
)

type Util struct {
}

func (u *Util) Rand(min int, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}
