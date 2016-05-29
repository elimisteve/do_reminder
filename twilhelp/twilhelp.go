// Steven Phillips / elimisteve
// 2016.05.28

package twilhelp

import (
	"log"
	"os"

	"github.com/subosito/twilio"
)

var (
	TwilioAccount = os.Getenv("TWILIO_ACCOUNT")
	TwilioKey     = os.Getenv("TWILIO_KEY")
	FromNumber    = os.Getenv("FROM_NUMBER")

	tc = twilio.NewClient(TwilioAccount, TwilioKey, nil)
)

func init() {
	if TwilioAccount == "" {
		log.Println("TWILIO_ACCOUNT not set")
	}
	if TwilioKey == "" {
		log.Println("TWILIO_KEY not set")
	}
	if FromNumber == "" {
		log.Println("FROM_NUMBER not set")
	}

	if len(FromNumber) == 10 {
		FromNumber = "+1" + FromNumber
	}
}

func SendSMS(toNumber, msg string) error {
	if len(toNumber) == 10 {
		toNumber = "+1" + toNumber
	}
	params := twilio.MessageParams{Body: msg}
	_, _, err := tc.Messages.Send(FromNumber, toNumber, params)
	return err
}
