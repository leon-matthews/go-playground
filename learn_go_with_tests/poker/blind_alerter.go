package poker

import (
	"fmt"
	"os"
	"time"
)

// BlindAlerter alerts users to blind increases after some time interval
type BlindAlerter interface {
	ScheduleAlert(duration time.Duration, amount int)
}

type BlindAlerterFunc func(duration time.Duration, amount int)

func (a BlindAlerterFunc) ScheduleAlert(duration time.Duration, amount int) {
	a(duration, amount)
}

// StdOutAlerter is called by BlindAlerterFunc, which implements BlindAlerter
func StdOutAlerter(duration time.Duration, amount int) {
	time.AfterFunc(duration, func() {
		fmt.Fprintf(os.Stdout, "Blind is now %d\n", amount)
	})
}
