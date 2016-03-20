package comm

import (
	"math/rand"
	"time"
)

func Delay(min, max time.Duration) time.Duration {
	return min + time.Duration(float64(max-min)*rand.Float64())
}
