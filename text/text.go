// Steven Phillips / elimisteve
// 2016.05.27

package main

import (
	"encoding/xml"
	"log"
	"os"
	"path/filepath"

	"github.com/subosito/twilio"
)

var (
	twilioAccount = os.Getenv("TWILIO_ACCOUNT")
	twilioKey     = os.Getenv("TWILIO_KEY")
	myNumber      = os.Getenv("FROM_NUMBER")

	tc = twilio.NewClient(twilioAccount, twilioKey, nil)
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Usage: %s <10-digit phone> <message>",
			filepath.Base(os.Args[0]))
	}

	toPhone := os.Args[1]
	msg := os.Args[2]

	if len(toPhone) == 10 {
		toPhone = "+1" + toPhone
	}

	if err := smsSend(toPhone, msg); err != nil {
		log.Fatalf("Error sending text to %s: %v\n", toPhone, err)
	}
}

func smsSend(recipient, replyBody string) error {
	params := twilio.MessageParams{Body: replyBody}
	_, _, err := tc.Messages.Send(myNumber, recipient, params)
	return err
}

func twilioResponse(s string) string {
	return xml.Header + "<Response>\n" + s + "\n</Response>"
}
