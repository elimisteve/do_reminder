// Steven Phillips / elimisteve
// 2016.05.27

package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/elimisteve/do_reminder/twilhelp"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Usage: %s <10-digit phone> <message>",
			filepath.Base(os.Args[0]))
	}

	toPhone := os.Args[1]
	msg := os.Args[2]

	if err := twilhelp.SendSMS(toPhone, msg); err != nil {
		log.Fatalf("Error sending text to %s: %v\n", toPhone, err)
	}
}
