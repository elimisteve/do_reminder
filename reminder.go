// Steve Phillips / elimisteve
// 2015.10.26

package main

import (
	"encoding/xml"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/codegangsta/martini"
	"github.com/elimisteve/do_reminder/remind"
	"github.com/elimisteve/do_reminder/twilhelp"
)

func init() {
	rand.Seed(remind.Now().Unix())
}

func main() {
	// Load DB
	options := &bolt.Options{Timeout: 2 * time.Second}
	dbPath := path.Join(os.Getenv("BOLT_PATH"), "reminder.db")
	db, err := bolt.Open(dbPath, 0600, options)
	if err != nil {
		log.Fatalf("Error opening bolt DB: %v", err)
	}
	defer db.Close()

	// Schedule all (non-cancelled) Reminders
	rems, err := remind.GetAllReminders(db)
	if err != nil {
		log.Fatalf("Error getting reminders: %v\n", err)
	}
	if err := rems.Schedule(db); err != nil {
		log.Fatal(err)
	}

	//
	// Router, etc
	//

	r := martini.NewRouter()
	m := martini.New()
	m.Use(martini.Logger())
	m.Use(martini.Recovery())
	m.Action(r.Handle)

	m.Map(db)

	r.Post("/sms", incomingSMS)

	m.Run()
}

func twilioResponse(s string) string {
	return xml.Header + "<Response>\n" + s + "\n</Response>"
}

var regexRemindMe = regexp.MustCompile(`^\s*[Rr]emind me to (.+?)\s*(@|at|around)\s*(\d?\d:\d\d)\s*(?:on)?\s*(today|tonight|tomorrow|\d?\d/\d?\d)?`)

func incomingSMS(db *bolt.DB, req *http.Request, log *log.Logger) string {
	now := remind.Now()
	from := req.FormValue("From")
	body := req.FormValue("Body")

	log.Printf("Incoming SMS: `%v: %v`", from, body)

	// Remind me to _ @ _

	description, nextRun, plusMinus, err := parseBody(body)
	if err != nil {
		log.Printf("Error parsing incoming message body: %v\n", err)
		return twilioResponse("")
	}

	// TODO(elimisteve): Make this configurable
	delta := 24 * time.Hour

	reminder := &remind.Reminder{
		Recipient:   from,
		Description: strings.ToUpper(description[0:1]) + description[1:],
		NextRun:     nextRun,
		Period:      delta,
		PlusMinus:   plusMinus,

		Raw:     body,
		Created: now,
	}

	err = reminder.Save(db)
	if err != nil {
		if err != nil {
			log.Printf("Error saving reminder %#v: %v\n", reminder, err)
		}

		err2 := twilhelp.SendSMS(from, "Error saving your reminder. Sorry!")
		if err2 != nil {
			log.Printf(`Error sending "sorry we couldn't save" msg: %v\n`, err)
		}

		return twilioResponse("")
	}

	err = reminder.Schedule(db)
	if err != nil {
		log.Printf("Error scheduling reminder %#v: %v\n", reminder, err)
		err2 := twilhelp.SendSMS(from, "Error scheduling your reminder. Sorry!")
		if err2 != nil {
			log.Printf("Error after successful parse but failed scheduling: %v\n", err2)
		}
		return twilioResponse("")
	}

	err = twilhelp.SendSMS(from, "Reminder successfully scheduled! Have a"+
		" great day :-)")
	if err != nil {
		log.Printf("Error from post-successful scheduling send: %v\n", err)
	}

	return twilioResponse("")
}

func parseBody(body string) (string, time.Time, time.Duration, error) {
	parts := regexRemindMe.FindStringSubmatch(body)
	if len(parts) < 4 {
		err := errors.New("Could not schedule your reminder. Be sure to" +
			" use military time (24-hour time) when saying something like," +
			"\n\nRemind me to take out the trash @ 18:00")
		log.Printf("Error sending after failed time parsing: %v\n", err)
		return "", time.Time{}, 0, err
	}

	// len(parts) >= 4

	// log.Printf("parts == %#v\n", parts)

	// parts[0] is the entire SMS message; ignore
	description := parts[1]
	around := (parts[2] == "around")
	nextRun, err := parseTime(parts[3], parts[4:])
	if err != nil {
		return "", time.Time{}, 0, err
	}

	// TODO(elimisteve): Use parts[5:] for more advanced features,
	// like reminders that aren't every day

	var plusMinus time.Duration
	if around {
		// TODO: Make configurable
		plusMinus = 60 * time.Minute
	}

	return description, nextRun, plusMinus, nil
}

func parseTime(t string, times []string) (time.Time, error) {
	when := strings.SplitN(t, ":", 2)
	hours, _ := strconv.Atoi(when[0])
	mins, _ := strconv.Atoi(when[1])

	var nextRun time.Time

	now := remind.Now()

	if len(times) == 0 || times[0] == "today" || times[0] == "tonight" || times[0] == "tomorrow" || times[0] == "" {
		nowHours, nowMins, nowSecs := now.Clock()
		today := now.
			Add(time.Duration(-nowHours) * time.Hour).
			Add(time.Duration(-nowMins) * time.Minute).
			Add(time.Duration(-nowSecs) * time.Second)

		nextRun = today.
			Add(time.Duration(hours) * time.Hour).
			Add(time.Duration(mins) * time.Minute)

		if nextRun.Before(now) || (len(times) > 0 && times[0] == "tomorrow") {
			// Tomorrow
			nextRun = nextRun.Add(24 * time.Hour)
		}

		return nextRun, nil
	}

	// Guaranteed: times[0] is of the form `\d?\d/\d?\d`

	monthDay := strings.SplitN(times[0], "/", 2)
	month, _ := strconv.Atoi(monthDay[0])
	day, _ := strconv.Atoi(monthDay[1])

	nextRun = time.Date(now.Year(), time.Month(month), day,
		hours, mins, 0, 0, remind.LosAngeles)

	if nextRun.Before(now) {
		// Next year
		nextRun = time.Date(now.Year()+1, time.Month(month), day,
			hours, mins, 0, 0, remind.LosAngeles)
	}

	return nextRun, nil
}
