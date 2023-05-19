package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_fixConfidentialiteRedressements_sansConfidentiel(t *testing.T) {
	ass := assert.New(t)
	redressements := []string{"a", "b", "c"}
	confidentiels := []string{"d"}
	expected := redressements
	newRedressements := fixConfidentialiteRedressements(redressements, confidentiels)

	ass.Equal(expected, newRedressements)
}

func Test_fixConfidentialiteRedressements_avecConfidentiel(t *testing.T) {
	ass := assert.New(t)
	redressements := []string{"a", "b", "c"}
	confidentiels := []string{"a"}
	expected := []string{"b", "c", "confidentiel"}
	newRedressements := fixConfidentialiteRedressements(redressements, confidentiels)

	ass.Equal(expected, newRedressements)
}

func Test_fixConfidentialiteRedressements_avecPlusieursConfidentiel(t *testing.T) {
	ass := assert.New(t)
	redressements := []string{"a", "b", "c"}
	confidentiels := []string{"a", "c"}
	expected := []string{"b", "confidentiel"}
	newRedressements := fixConfidentialiteRedressements(redressements, confidentiels)
	ass.Equal(expected, newRedressements)
}
