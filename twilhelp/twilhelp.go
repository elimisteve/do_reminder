// Steven Phillips / elimisteve
// 2016.05.28

package twilhelp

import (
	"log"
	"os"

	"github.com/subosito/twilio"
)

var (
	twilioAccount = os.Getenv("TWILIO_ACCOUNT")
	twilioKey     = os.Getenv("TWILIO_KEY")
	fromNumber    = os.Getenv("FROM_NUMBER")

	tc = twilio.NewClient(twilioAccount, twilioKey, nil)
)

func init() {
	if twilioAccount == "" {
		log.Println("TWILIO_ACCOUNT not set")
	}
	if twilioKey == "" {
		log.Println("TWILIO_KEY not set")
	}
	if fromNumber == "" {
		log.Println("FROM_NUMBER not set")
	}
}

func SendSMS(toNumber, replyBody string) error {
	if len(toNumber) == 10 {
		toNumber = "+1" + toNumber
	}
	params := twilio.MessageParams{Body: replyBody}
	_, _, err := tc.Messages.Send(fromNumber, toNumber, params)
	return err
}
