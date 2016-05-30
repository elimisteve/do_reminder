// Steven Phillips / elimisteve
// 2016.05.27

package main

import (
	"testing"
	"time"

	"github.com/elimisteve/do_reminder/remind"
	"github.com/stretchr/testify/assert"
)

func TestRegex(t *testing.T) {
	nowYear, nowMonth, nowDay := remind.Now().Date()

	tests := []struct {
		body string
		rem  remind.Reminder
	}{
		{
			"Remind me to buy milk at 14:45 tomorrow",
			remind.Reminder{
				Description: "Buy milk",
				NextRun: time.Date(nowYear, nowMonth, nowDay+1,
					14, 45, 0, 0, remind.LosAngeles),
			},
		},
		{
			"Remind me to do $tuFF at 23:59 today",
			remind.Reminder{
				Description: "Do $tuFF",
				NextRun: time.Date(nowYear, nowMonth, nowDay,
					23, 59, 0, 0, remind.LosAngeles),
			},
		},
		{
			"Remind me to do  whatever at 23:59",
			remind.Reminder{
				Description: "Do  whatever",
				NextRun: time.Date(nowYear, nowMonth, nowDay,
					23, 59, 0, 0, remind.LosAngeles),
			},
		},
		{
			"  remind me to write_GO/code!!.(?)  @ 23:59 on  12/09",
			remind.Reminder{
				Description: "Write_GO/code!!.(?)",
				NextRun: time.Date(nowYear, 12, 9,
					23, 59, 0, 0, remind.LosAngeles),
			},
		},
		{
			"Remind me to take out the trash @ 18:00",
			remind.Reminder{
				Description: "Take out the trash",
				NextRun: time.Date(nowYear, nowMonth, nowDay,
					18, 00, 0, 0, remind.LosAngeles),
			},
		},
		{
			"Remind me to take out the trash @ 18:00 starting 1/1",
			remind.Reminder{
				Description: "Take out the trash",
				NextRun: time.Date(nowYear+1, 1, 1,
					18, 00, 0, 0, remind.LosAngeles),
				Period: 24 * time.Hour,
			},
		},
		{
			"Remind me to take out the trash @ 18:00 on 1/1 daily",
			remind.Reminder{
				Description: "Take out the trash",
				NextRun: time.Date(nowYear+1, 1, 1,
					18, 00, 0, 0, remind.LosAngeles),
				Period: 24 * time.Hour,
			},
		},
	}

	for _, test := range tests {
		r, err := parseReminder("", test.body)
		if err != nil {
			t.Errorf("Error parsing `%s`: %v", test.body, err)
			continue
		}

		assert.Equal(t, r.Description, test.rem.Description, "Description is wrong")
		assert.Equal(t, r.NextRun, test.rem.NextRun, "NextRun is wrong")
		assert.Equal(t, r.Period, test.rem.Period, "Period is wrong")
		assert.Equal(t, r.PlusMinus, time.Duration(0), "PlusMinus is wrong")

		r, _ = parseReminder("", test.body+" daily")
		assert.Equal(t, r.Period, 24*time.Hour, "Period is wrong (daily)")
	}
}
