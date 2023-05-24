package factory

import (
	"github.com/jaswdr/faker"
)

var fake faker.Faker

func init() {
	fake = faker.New()
}
