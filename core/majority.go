package core

import (
	"math"
)

// MajorityCount returns 2f + 1 members
func MajorityCount(validatorCount int) int {
	// n=3f+1 -> f=floor((n-1)3) -> m=n-f -> m=ceil((2n+1)/3)
	return int(math.Ceil(float64(2*validatorCount+1) / 3))
}
