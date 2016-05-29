// Steven Phillips / elimisteve
// 2016.05.28

package remind

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(Now().Unix())
}

// IntervalSMS sends to toNumber each message from msgs throughout the
// day, from start to finish, where this interval is broken into
// len(msgs) pieces.
func IntervalSMS(toNumber string, msgs []string, start, finish time.Time) ([]*Reminder, error) {
	if len(toNumber) < 10 {
		return nil, fmt.Errorf("Phone number '%s' is too short", toNumber)
	}
	if !start.Before(finish) {
		return nil, fmt.Errorf("start time %s must be before finish %s", start, finish)
	}
	if len(msgs) == 0 {
		return nil, errors.New("No messages included")
	}

	// E.g., 12pm to 5pm. If len(msgs) == 4, oneChunk == 1.25h
	oneChunk := finish.Sub(start) / time.Duration(len(msgs))

	reminders := make([]*Reminder, len(msgs))

	nMsgs := len(msgs)
	for i := 0; i < nMsgs; i++ {
		// Randomly choose message
		ndx := rand.Intn(len(msgs))
		msg := msgs[ndx]

		// Remove message so it can't be reused
		msgs = append(msgs[:ndx], msgs[ndx+1:]...)

		// Start + jump to beginning of correct interval + add random
		// duration between 0 and oneChuck
		nextRun := start.
			Add(time.Duration(i) * oneChunk).
			Add(randPositiveDur(oneChunk))

		rem := Reminder{
			Recipient:   toNumber,
			Description: msg,
			NextRun:     nextRun,
			Created:     Now(),
		}

		reminders[i] = &rem
	}

	return reminders, nil
}

func randPositiveDur(d time.Duration) time.Duration {
	randDur := RandDuration(d)
	if randDur < 0 {
		randDur = -randDur
	}
	return randDur
}
