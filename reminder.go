// Steve Phillips / elimisteve
// 2015.10.26

package main

import (
	"encoding/xml"
	"log"
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

var regexRemindMe = regexp.MustCompile(`^[Rr]emind me to (.+?)\s*@\s*(\d\d:\d\d)\s*(today|tonight|tomorrow|\d\d?/\d\d?)?`)

func incomingSMS(tc *twilio.Client, req *http.Request, log *log.Logger) string {
	from := req.FormValue("From")
	body := req.FormValue("Body")

	log.Printf("Incoming SMS: `%v: %v`", from, body)

	// Remind me to _ @ _
	parts := regexRemindMe.FindStringSubmatch(body)
	if len(parts) < 2 {
		err := smsReply(from, "Could not schedule your reminder. Be sure to"+
			" use military time (24-hour time) when saying something like,"+
			"\n\nRemind me to take out the trash @ 18:00")
		if err != nil {
			log.Printf("Error sending after failed time parsing: %v\n", err)
		}
		return twilioResponse("")
	}

	// len(parts) >= 2

	// parts[0] is the entire SMS message; ignore
	description := parts[1]
	delay := parseTime(parts[2])

	// TODO(elimisteve): Make this configurable
	epsilon := 24 * time.Hour

	reminder := &Reminder{
		Raw:          body,
		Recipient:    from,
		Description:  description,
		InitialDelay: delay,
		Period:       epsilon,
	}

	err := reminder.Schedule()
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

func parseTime(t string) time.Duration {
	when := strings.SplitN(t, ":", 2)
	hours, _ := strconv.Atoi(when[0])
	mins, _ := strconv.Atoi(when[1])

	now := Now()
	nowHours, nowMins, nowSecs := now.Clock()
	today := now.
		Add(time.Duration(-nowHours) * time.Hour).
		Add(time.Duration(-nowMins) * time.Minute).
		Add(time.Duration(-nowSecs) * time.Second)

	nextRun := today.
		Add(time.Duration(hours) * time.Hour).
		Add(time.Duration(mins) * time.Minute)

	if nextRun.Before(now) {
		nextRun = nextRun.Add(24 * time.Hour)
	}

	return nextRun.Sub(now)
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
	Raw          string
	Recipient    string
	Description  string
	InitialDelay time.Duration
	Period       time.Duration
}

// Schedule reminds recipient to do description starting at nextRun then
// every epsilon after that
func (r *Reminder) Schedule() error {
	r.Description = strings.ToUpper(r.Description[0:1]) + r.Description[1:]

	go func() {
		log.Printf("Scheduling *Reminder `%#v`\n", r)

		time.Sleep(r.InitialDelay)
		for {
			log.Printf("Texting `%s` to remind him/her to `%s` starting now then every `%s` after that\n",
				r.Recipient, r.Description, r.Period)

			err := smsReply(r.Recipient, r.Description)
			if err != nil {
				log.Printf("Error reminding `%v` to `%v`\n", r.Recipient, r.Description)
			}

			<-time.Tick(r.Period)
		}
	}()

	return nil
}
