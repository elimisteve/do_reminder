// Steven Phillips / elimisteve
// 2016.05.28

package remind

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRandDuration(t *testing.T) {
	durs := []time.Duration{
		-5 * time.Second,
		0 * time.Second,
		2 * time.Second,
		1 * time.Minute,
		10 * time.Hour,
	}

	for _, d := range durs {
		for i := 0; i < 100; i++ {
			r := RandDuration(d)
			if d < 0 {
				d *= -1
			}
			assert.True(t, -d.Seconds() <= r.Seconds() && r.Seconds() <= d.Seconds(),
				"Duration produced that was outside of the desired range")
		}
	}
}
