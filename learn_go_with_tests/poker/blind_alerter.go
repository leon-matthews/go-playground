package poker

import (
	"fmt"
	"os"
	"time"
)

// Alerter schedules alerts for points in the future
type Alerter interface {
	// Schedule an alert for the change to the blind to amount dollars
	Schedule(at time.Duration, amount int)
}

// AlerterFunc implements [Alerter] from plain function
type AlerterFunc func(duration time.Duration, amount int)

func (a AlerterFunc) Schedule(duration time.Duration, amount int) {
	a(duration, amount)
}

// StdOutAlerter simply prints alert out to stdout after given duration
func StdOutAlerter(duration time.Duration, amount int) {
	time.AfterFunc(duration, func() {
		fmt.Fprintf(os.Stdout, "Blind is now %d\n", amount)
	})
}
