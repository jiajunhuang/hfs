package selection

import (
	"math/rand"
)

/*
package selection provide chunk selection algorithms.
*/

// Random random select min(len(avalable), num) elems in avalable and return
func Random(avalable []string, itself string, num int32) []string {
	if num <= 1 {
		return []string{}
	}

	last := len(avalable) - 1
	for i, a := range avalable {
		if a == itself {
			avalable[i], avalable[last] = avalable[last], avalable[i] // swap
			avalable = avalable[:last]                                // ignore itself
			break
		}
	}

	// shuffle
	rand.Shuffle(len(avalable), func(i, j int) {
		avalable[i], avalable[j] = avalable[j], avalable[i]
	})

	min := int(num)
	if len(avalable) < min {
		min = len(avalable)
	}

	return avalable[:min]
}
