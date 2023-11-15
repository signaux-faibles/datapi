package stats

import (
	"time"

	"github.com/jaswdr/faker"
)

var fakeTU faker.Faker
var segments []string
var usernames []string

func init() {
	fakeTU = faker.New()
	segments = []string{"bdf", "ddfip", "urssaf", "sf"}
	usernames = []string{
		fakeTU.Internet().Email(),
		fakeTU.Internet().Email(),
		fakeTU.Internet().Email(),
		fakeTU.Internet().Email(),
		fakeTU.Internet().Email(),
		fakeTU.Internet().Email(),
		fakeTU.Internet().Email(),
		fakeTU.Internet().Email(),
		fakeTU.Internet().Email(),
		fakeTU.Internet().Email(),
	}
}

func createFakeActivitesChan[A any](activitiesNumber int, newActivite func() A) (A, chan row[A]) {
	r := make(chan row[A])
	activite := newActivite()
	go func() {
		defer close(r)
		r <- newRow(activite)
		for i := 1; i < activitiesNumber; i++ {
			r <- newRow(newActivite())
		}
	}()
	return activite, r
}

type testline struct {
	valString  string    `col:"chaîne de caractères" size:"25"`
	valDay     time.Time `col:"jour" size:"12"`
	valInstant time.Time `col:"instant" size:"16"`
	valInt     int       `col:"entier" size:"16"`
}

func createTestChannel(items ...testline) chan row[testline] {
	r := make(chan row[testline])
	go func() {
		defer close(r)
		for _, current := range items {
			r <- newRow(current)
		}
	}()
	return r
}

func createFakeActiviteUtilisateur() activiteParUtilisateur {
	var visites, actions int
	visites = fakeTU.IntBetween(1, 99)
	actions = visites * fakeTU.IntBetween(1, 11)
	return activiteParUtilisateur{
		username: fakeTU.RandomStringElement(usernames),
		actions:  actions,
		visites:  visites,
		segment:  fakeTU.RandomStringElement(segments),
	}
}

func createTestLine() testline {
	now := time.Now()
	t := fakeTU.Time().TimeBetween(now.AddDate(1, 0, 0), now)

	return testline{
		valString:  fakeTU.Lorem().Word(),
		valDay:     truncateToDay(t),
		valInstant: t,
		valInt:     fakeTU.IntBetween(1, 999),
	}
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
