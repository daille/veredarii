package global

import (
	"fmt"
	"math/rand"
)

func RandomInt(min, max int) (int, error) {
	if min > max {
		return 0, fmt.Errorf("min cannot be greater than max")
	}
	return rand.Intn(max-min+1) + min, nil
}
