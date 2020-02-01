// Steven Phillips / elimisteve
// 2016.05.28

package remind

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIntervalSMS(t *testing.T) {
	now := Now()

	tests := []struct {
		start, finish time.Time
		msgs          []string
	}{
		{now, now.Add(3 * time.Minute), []string{
			"Message 0", "Message 1", "Message 2",
		}},
		{now, now.Add(1 * time.Minute), []string{
			"Message 0 or 4",
		}},
		{now, now.Add(3 * time.Minute), []string{
			"Message 0 or 5",
		}},
	}

	for _, test := range tests {
		rems, err := IntervalSMS("0123456789", test.msgs, test.start, test.finish)
		if err != nil {
			t.Errorf("Error from IntervalSMS: %v", err)
			continue
		}

		// Every reminder should be scheduled within window (between
		// start and finish)

		for i, rem := range rems {
			t.Logf("rem[%d] == %#v\n", i, rem)

			assert.WithinDuration(t,
				test.start,
				rem.NextRun,
				test.finish.Sub(test.start),
				fmt.Sprintf(
					"Reminder scheduled for %s, should be between %s and %s",
					rem.NextRun, test.start, test.finish))
		}

		oneChunk := test.finish.Sub(test.start) / time.Duration(len(test.msgs))

		t.Logf("oneChunk == %v", oneChunk)

		// Each pair of messages should be at most oneChunk*2 apart

		for i := 0; i < len(rems)-1; i++ {
			// Adjacent reminders _should_ be closer than oneChunk * 2
			assert.WithinDuration(
				t,
				rems[i].NextRun,
				rems[i+1].NextRun,
				oneChunk*time.Duration(2),
				fmt.Sprintf("Reminders %v and %v scheduled too far apart (%v)",
					rems[i],
					rems[i+1],
					time.Duration(2)*rems[i+1].NextRun.Sub(rems[i].NextRun)),
			)
		}
	}

}
