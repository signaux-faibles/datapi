package stats

import (
	"strconv"
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

func createFakeActiviteJour() activiteParJour {
	r := fakeTU.IntBetween(1, 99)
	f := fakeTU.IntBetween(3, 99)
	a := fakeTU.IntBetween(3, 99) + r + f
	return activiteParJour{
		jour:       fakeTU.Time().TimeBetween(time.Now().AddDate(0, -3, 0), time.Now()),
		username:   fakeTU.RandomStringElement(usernames),
		actions:    a,
		recherches: r,
		fiches:     f,
		segment:    fakeTU.RandomStringElement(segments),
	}
}

func createStringChannel(items ...string) chan row[string] {
	r := make(chan row[string])
	go func() {
		defer close(r)
		for _, i := range items {
			r <- newRow(i)
		}
	}()
	return r
}

func createFakeActiviteUtilisateur() activiteParUtilisateur {
	var visites, actions int
	visites = fakeTU.IntBetween(1, 99)
	actions = visites * fakeTU.IntBetween(1, 11)
	return activiteParUtilisateur{
		username: fakeTU.Internet().Email(),
		actions:  strconv.Itoa(actions),
		visites:  strconv.Itoa(visites),
		segment:  fakeTU.RandomStringElement(segments),
	}
}
