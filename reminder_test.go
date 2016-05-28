// Steven Phillips / elimisteve
// 2016.05.27

package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRegex(t *testing.T) {
	nowYear, nowMonth, nowDay := Now().Date()

	tests := []struct {
		body string
		rem  Reminder
	}{
		{
			"Remind me to buy milk at 14:45 tomorrow",
			Reminder{
				Description: "buy milk",
				NextRun: time.Date(nowYear, nowMonth, nowDay+1,
					14, 45, 0, 0, LosAngeles),
			},
		},
		{
			"Remind me to do $tuFF at 23:59 today",
			Reminder{
				Description: "do $tuFF",
				NextRun: time.Date(nowYear, nowMonth, nowDay,
					23, 59, 0, 0, LosAngeles),
			},
		},
		{
			"Remind me to do  whatever at 23:59",
			Reminder{
				Description: "do  whatever",
				NextRun: time.Date(nowYear, nowMonth, nowDay,
					23, 59, 0, 0, LosAngeles),
			},
		},
		{
			"  remind me to write_GO/code!!.(?)  @ 23:59 on  12/09",
			Reminder{
				Description: "write_GO/code!!.(?)",
				NextRun: time.Date(nowYear, 12, 9,
					23, 59, 0, 0, LosAngeles),
			},
		},
	}

	for _, test := range tests {
		descrip, nextRun, _, err := parseBody("", test.body)
		if err != nil {
			t.Errorf("Error parsing `%s`: %v", test.body, err)
			continue
		}

		assert.Equal(t, descrip, test.rem.Description, "Description is wrong")
		assert.Equal(t, nextRun, test.rem.NextRun, "NextRun is wrong")
	}
}

func TestRandDur(t *testing.T) {
	durs := []time.Duration{
		2 * time.Second,
		1 * time.Minute,
		10 * time.Hour,
	}

	for _, d := range durs {
		for i := 0; i < 100; i++ {
			r := randDur(d)
			assert.True(t, -d.Seconds() <= r.Seconds() && r.Seconds() <= d.Seconds(),
				"Duration produced that was outside of the desired range")
		}
	}
}
