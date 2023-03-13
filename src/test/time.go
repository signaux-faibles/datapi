package test

import (
	"bou.ke/monkey"
	"time"
)

var TuTime = time.Date(2023, 03, 10, 17, 41, 58, 651387237, time.UTC)

func FakeTime(t time.Time) {
	monkey.Patch(time.Now, func() time.Time {
		return t
	})
}
