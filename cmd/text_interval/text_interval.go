// Steven Phillips / elimisteve
// 2016.05.28

package main

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/elimisteve/do_reminder/remind"
)

var regexTime = regexp.MustCompile(`^\d?\d:\d\d$`)

func main() {
	if len(os.Args) < 5 {
		log.Fatalf("Usage: %s <to_number> <start_time> <finish_time> <msg 1> [<msg 2> ...]\n", filepath.Base(os.Args[0]))
	}

	toNumber := os.Args[1]
	startStr := os.Args[2]  // \d?\d:\d\d
	finishStr := os.Args[3] // \d?\d:\d\d
	messages := os.Args[4:]

	if !regexTime.MatchString(startStr) {
		log.Fatalf("Start time must be of the form hh:mm; you typed '%s'", startStr)
	}
	if !regexTime.MatchString(finishStr) {
		log.Fatalf("Finish time must be of the form hh:mm; you typed '%s'", finishStr)
	}

	startHour, _ := strconv.Atoi(startStr[:2])
	startMinute, _ := strconv.Atoi(startStr[3:])

	finishHour, _ := strconv.Atoi(finishStr[:2])
	finishMinute, _ := strconv.Atoi(finishStr[3:])

	now := remind.Now()

	year, month, day := now.Date()
	start := time.Date(year, month, day, startHour, startMinute,
		0, 0, remind.LosAngeles)
	finish := time.Date(year, month, day, finishHour, finishMinute,
		0, 0, remind.LosAngeles)

	reminders, err := remind.IntervalSMS(toNumber, messages, start, finish)
	if err != nil {
		log.Fatalf("Error from IntervalSMS: %v\n", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(reminders))

	for _, rem := range reminders {
		go func(rem *remind.Reminder) {
			defer wg.Done()

			if rem.NextRun.Before(now) {
				log.Printf("Reminder `%s` scheduled for the past (%s ago); not running\n",
					rem.Description, now.Sub(rem.NextRun))
				return
			}

			sleep := rem.NextRun.Sub(now)
			log.Printf("Reminder `%s` will run in %s\n", rem.Description, sleep)
			time.Sleep(sleep)

			if err := rem.SendSMS(); err != nil {
				log.Printf("Error sending Reminder `%v` to %s: %v\n",
					rem.Description, rem.Recipient, err)
				return
			}
			log.Printf("Reminder `%s` sent successfully\n", rem.Description)
		}(rem)
	}

	wg.Wait()

	log.Printf("All %d Reminders finished; exiting\n", len(reminders))
}
