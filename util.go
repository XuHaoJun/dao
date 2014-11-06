package dao

import (
	"math/rand"
	"time"
)

type Util struct {
}

func RandIntnRange(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Intn(max-min) + min
}

// FIXME
// each rand will seed again, very slow....
func (u *Util) Rand(min int, max int) int {
	return RandIntnRange(min, max)
}
