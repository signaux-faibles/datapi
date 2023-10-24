package stats

import (
	"strconv"

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

type dummy struct {
	value string `col:"valeur" size:"16"`
}

func createStringChannel(items ...string) chan row[dummy] {
	r := make(chan row[dummy])
	go func() {
		defer close(r)
		for _, i := range items {
			r <- newRow(dummy{i})
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
