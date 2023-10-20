//go:build integration

package stats

import (
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

func Test_selectAny(t *testing.T) {
	possibleValues := []string{fake.Lorem().Word(), fake.Lorem().Word(), fake.Lorem().Word()}
	selecta := newDummyStringSelector(possibleValues...)
	actualChan := fetchAny(api, "", selecta)
	for current := range actualChan {
		assert.Contains(t, possibleValues, current)
	}
}

type dummyStringSelecta struct {
	values []string
}

func (d dummyStringSelecta) sql() string {
	return ""
}

func (d dummyStringSelecta) sqlArgs() []any {
	return nil
}

func (d dummyStringSelecta) toItem(_ pgx.Rows) (string, error) {
	return fake.RandomStringElement(d.values), nil
}

func newDummyStringSelector(values ...string) dummyStringSelecta {
	return dummyStringSelecta{values: values}
}
