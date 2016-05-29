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
				Description: "buy milk",
				NextRun: time.Date(nowYear, nowMonth, nowDay+1,
					14, 45, 0, 0, LosAngeles),
			},
		},
		{
			"Remind me to do $tuFF at 23:59 today",
			remind.Reminder{
				Description: "do $tuFF",
				NextRun: time.Date(nowYear, nowMonth, nowDay,
					23, 59, 0, 0, LosAngeles),
			},
		},
		{
			"Remind me to do  whatever at 23:59",
			remind.Reminder{
				Description: "do  whatever",
				NextRun: time.Date(nowYear, nowMonth, nowDay,
					23, 59, 0, 0, LosAngeles),
			},
		},
		{
			"  remind me to write_GO/code!!.(?)  @ 23:59 on  12/09",
			remind.Reminder{
				Description: "write_GO/code!!.(?)",
				NextRun: time.Date(nowYear, 12, 9,
					23, 59, 0, 0, LosAngeles),
			},
		},
	}

	for _, test := range tests {
		descrip, nextRun, _, err := parseBody(test.body)
		if err != nil {
			t.Errorf("Error parsing `%s`: %v", test.body, err)
			continue
		}

		assert.Equal(t, descrip, test.rem.Description, "Description is wrong")
		assert.Equal(t, nextRun, test.rem.NextRun, "NextRun is wrong")
	}
}
