// Steven Phillips / elimisteve
// 2016.05.28

package twilhelp

import (
	"math/rand"
	"time"
)

// RandDuration returns a random duration r where -d <= r <= d. Works
// to 1-second resolution.
func RandDuration(d time.Duration) time.Duration {
	if d == 0 {
		return d
	}
	if d < 0 {
		d *= -1
	}

	randSecs := rand.Intn(int(d.Seconds()))
	if rand.Intn(2) == 0 { // 50% chance
		randSecs *= -1
	}

	return time.Duration(randSecs) * time.Second
}
