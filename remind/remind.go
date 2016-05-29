// Steven Phillips / elimisteve
// 2016.05.28

package remind

import (
	"fmt"
	"log"
	"time"

	"github.com/elimisteve/do_reminder/twilhelp"
)

type Reminder struct {
	Recipient   string
	Description string
	NextRun     time.Time
	Period      time.Duration
	PlusMinus   time.Duration

	Raw     string
	Created time.Time
}

// Schedule reminds r.Recipient to do r.Description starting at
// r.NextRun, then every r.Period +/- r.PlusMinus after that.
func (r *Reminder) Schedule() error {
	log.Printf("Scheduling *Reminder `%#v`\n", r)

	if r.NextRun.Before(Now()) {
		return fmt.Errorf("Reminder %#v's next run already passed", r)
	}

	if r.PlusMinus < 0 {
		r.PlusMinus *= -1
	}
	nextRun := r.NextRun.Add(RandDuration(r.PlusMinus))

	// Sleep till the next run is here
	dur := max(nextRun.Sub(Now()), 0)

	time.Sleep(dur)
	for {
		log.Printf("Texting `%s` to remind him/her to `%s` starting now then every ~%s after that\n",
			r.Recipient, r.Description, r.Period)

		err := twilhelp.SendSMS(r.Recipient, r.Description)
		if err != nil {
			log.Printf("Error reminding `%v` to `%v`: %v\n", r.Recipient,
				r.Description, err)
		}

		if r.Period == 0 && r.PlusMinus == 0 {
			log.Printf("Reminder %#v should only run once; exiting\n", r)
			return nil
		}

		// TODO: Prevent drift. Right now there's nothing stopping
		// the time at which a reminder runs from drifting 60 mins
		// every single time!
		sleep := r.Period + RandDuration(r.PlusMinus)
		log.Printf("Text to %s, `%s`, sending again in %s (period: %s)\n",
			r.Recipient, r.Description, sleep, r.Period)
		time.Sleep(max(sleep, -sleep))
	}

	return nil
}

func max(n, m time.Duration) time.Duration {
	if n > m {
		return n
	}
	return m
}
