// Steve Phillips / elimisteve
// 2015.10.26

package main

import (
	"encoding/xml"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/codegangsta/martini"
	"github.com/subosito/twilio"
)

var (
	twilioAccount = os.Getenv("TWILIO_ACCOUNT")
	twilioKey     = os.Getenv("TWILIO_KEY")
	myNumber      = os.Getenv("FROM_NUMBER")

	tc = twilio.NewClient(twilioAccount, twilioKey, nil)

	// TODO(elimisteve): Make timezone user-specific
	LosAngeles, _ = time.LoadLocation("America/Los_Angeles")
)

func main() {
	rand.Seed(time.Now().Unix())

	r := martini.NewRouter()
	m := martini.New()
	m.Use(martini.Logger())
	m.Use(martini.Recovery())
	m.Action(r.Handle)

	m.Map(tc)

	r.Post("/sms", incomingSMS)

	m.Run()
}

func twilioResponse(s string) string {
	return xml.Header + "<Response>\n" + s + "\n</Response>"
}

var regexRemindMe = regexp.MustCompile(`^\s*[Rr]emind me to (.+?)\s*(@|at|around)\s*(\d?\d:\d\d)\s*(?:on)?\s*(today|tonight|tomorrow|\d?\d/\d?\d)?`)

func incomingSMS(tc *twilio.Client, req *http.Request, log *log.Logger) string {
	now := Now()
	from := req.FormValue("From")
	body := req.FormValue("Body")

	log.Printf("Incoming SMS: `%v: %v`", from, body)

	// Remind me to _ @ _

	description, nextRun, plusMinus, err := parseBody(from, body)
	if err != nil {
		log.Printf("Error parsing incoming message body: %v\n", err)
		return twilioResponse("")
	}

	// TODO(elimisteve): Make this configurable
	delta := 24 * time.Hour

	reminder := &Reminder{
		Recipient:   from,
		Description: strings.ToUpper(description[0:1]) + description[1:],
		NextRun:     nextRun,
		Period:      delta,
		PlusMinus:   plusMinus,

		Raw:     body,
		Created: now,
	}

	err = reminder.Schedule()
	if err != nil {
		err2 := smsReply(from, "Error scheduling your reminder. Sorry!")
		if err2 != nil {
			log.Printf("Error after successful parse but failed scheduling: %v\n", err2)
		}
		return twilioResponse("")
	}

	err = smsReply(from, "Reminder successfully scheduled! Have a great day :-)")
	if err != nil {
		log.Printf("Error from post-successful scheduling send: %v\n", err)
	}

	return twilioResponse("")
}

func parseBody(from, body string) (string, time.Time, time.Duration, error) {
	parts := regexRemindMe.FindStringSubmatch(body)
	if len(parts) < 3 {
		err := smsReply(from, "Could not schedule your reminder. Be sure to"+
			" use military time (24-hour time) when saying something like,"+
			"\n\nRemind me to take out the trash @ 18:00")
		if err != nil {
			log.Printf("Error sending after failed time parsing: %v\n", err)
		}
		return "", time.Time{}, 0, err
	}

	// len(parts) >= 3

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

	now := Now()

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
		hours, mins, 0, 0, LosAngeles)

	if nextRun.Before(now) {
		// Next year
		nextRun = time.Date(now.Year()+1, time.Month(month), day,
			hours, mins, 0, 0, LosAngeles)
	}

	return nextRun, nil
}

func Now() time.Time {
	return time.Now().In(LosAngeles).Round(time.Second)
}

func smsReply(recipient, replyBody string) error {
	params := twilio.MessageParams{Body: replyBody}
	_, _, err := tc.Messages.Send(myNumber, recipient, params)
	return err
}

//
// Types
//

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
	go func() {
		log.Printf("Scheduling *Reminder `%#v`\n", r)

		if r.PlusMinus < 0 {
			r.PlusMinus *= -1
		}

		nextRun := r.NextRun.Add(randDur(r.PlusMinus))

		// Sleep till the next run is here
		dur := max(nextRun.Sub(Now()), 0)

		time.Sleep(dur)
		for {
			log.Printf("Texting `%s` to remind him/her to `%s` starting now then every ~%s after that\n",
				r.Recipient, r.Description, r.Period)

			err := smsReply(r.Recipient, r.Description)
			if err != nil {
				log.Printf("Error reminding `%v` to `%v`\n", r.Recipient, r.Description)
			}

			// TODO: Prevent drift. Right now there's nothing stopping
			// the time at which a reminder runs from drifting 60 mins
			// every single time!
			sleep := r.Period + randDur(r.PlusMinus)
			log.Printf("Text to %s, `%s`, sending again in %s (period: %s)\n",
				r.Recipient, r.Description, sleep, r.Period)
			time.Sleep(max(sleep, 0))
		}
	}()

	return nil
}

// randDur returns a random duration r where -d <= r <= d
func randDur(d time.Duration) time.Duration {
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

func max(n, m time.Duration) time.Duration {
	if n > m {
		return n
	}
	return m
}
