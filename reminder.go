// Steve Phillips / elimisteve
// 2015.10.26

package main

import (
	"encoding/xml"
	"log"
	"net/http"
	"os"

	"github.com/codegangsta/martini"
	"github.com/subosito/twilio"
)

var (
	twilioAccount = os.Getenv("TWILIO_ACCOUNT")
	twilioKey     = os.Getenv("TWILIO_KEY")
	myNumber      = os.Getenv("FROM_NUMBER")
)

func main() {
	tc := twilio.NewClient(twilioAccount, twilioKey, nil)

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

func incomingSMS(tc *twilio.Client, req *http.Request, log *log.Logger) string {
	from := req.FormValue("From")
	body := req.FormValue("Body")

	log.Printf("Incoming SMS: `%v: %v`", from, body)

	replyBody := "Received this text from you: `" + body + "`"

	params := twilio.MessageParams{Body: replyBody}
	_, _, err := tc.Messages.Send(myNumber, from, params)
	if err != nil {
		log.Printf("Error sending SMS from Twilio (%v) to %v: `%v`\n", myNumber, from, err)
	} else {
		log.Printf("Successfully sent this SMS to `%s`: %s", from, replyBody)
	}

	return twilioResponse("")
}
