// Steven Phillips / elimisteve
// 2016.05.28

package twilhelp

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

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

func SendSMS(toNumberOrig, msg string) error {
	toNumber := cleanNumber(toNumberOrig)
	fmt.Printf("Cleaned: %s => %s\n", toNumberOrig, toNumber)
	params := twilio.MessageParams{Body: msg}
	_, _, err := tc.Messages.Send(FromNumber, toNumber, params)
	return err
}

var reNumber = regexp.MustCompile(`\d+`)

func cleanNumber(toNumberOrig string) string {
	digits := reNumber.FindAllString(toNumberOrig, -1)
	num := strings.Join(digits, "")
	if len(num) == 10 {
		num = "+1" + num
	}
	if !strings.HasPrefix(num, "+") {
		num = "+" + num
	}
	return num
}
