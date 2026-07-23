package humanise

import "time"

// Age returns the age in whole years at today for someone born on born.
//
// Age is counted the way people count it: a year is only claimed once its
// birthday has arrived. Someone born on 29 February ticks over on 1 March in
// non-leap years, which falls out of the day comparison without a special case.
// A born date in the future yields a negative age.
func Age(born, today time.Time) int {
	years := today.Year() - born.Year()

	// Undo the last year if this year's birthday has not yet arrived.
	if today.Month() < born.Month() ||
		(today.Month() == born.Month() && today.Day() < born.Day()) {
		years--
	}
	return years
}
