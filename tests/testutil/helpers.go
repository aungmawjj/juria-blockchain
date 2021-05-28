// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package testutil

import (
	"fmt"
	"math/rand"
	"time"
)

// Sleep print duration and call time.Sleep
func Sleep(d time.Duration) {
	fmt.Printf("Wait for %s\n", d)
	time.Sleep(d)
}

func PickUniqueRandoms(total, count int) []int {
	rand.Seed(time.Now().Unix())
	unique := make(map[int]struct{}, count)
	for len(unique) < count {
		unique[rand.Intn(total)] = struct{}{}
	}
	ret := make([]int, 0, count)
	for v := range unique {
		ret = append(ret, v)
	}
	return ret
}
