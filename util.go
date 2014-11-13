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

func RandIntnRangeInt63(min int64, max int64) int64 {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Int63n(max-min) + min
}

// FIXME
// each rand will seed again, very slow....
func (u *Util) Rand(min int, max int) int {
	return RandIntnRange(min, max)
}

func (u *Util) RandInt63(min int64, max int64) int64 {
	return RandIntnRangeInt63(min, max)
}
