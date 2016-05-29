// Steven Phillips / elimisteve
// 2016.05.28

package twilhelp

import "time"

var (
	// TODO(elimisteve): Make timezone user-specific
	LosAngeles, _ = time.LoadLocation("America/Los_Angeles")
)

func Now() time.Time {
	return time.Now().In(LosAngeles).Round(time.Second)
}
