package random

import (
	"math/rand/v2"
	"time"
)

type randSource struct{}

func (rsrc randSource) Uint64() uint64 {
	return uint64(time.Now().UnixNano())
}

func RandomNumRange(min, max int64) int64 {
	randgen := rand.New(randSource{})
	num := randgen.Int64N(max-min) + min

	return num
}
