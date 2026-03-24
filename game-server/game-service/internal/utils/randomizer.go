package utils

import "math/rand/v2"

func GenRandomBetween(min, max int) int {
	if max <= min {
		panic("max must be greater than min")
	}
	return min + rand.IntN(max-min)
}
