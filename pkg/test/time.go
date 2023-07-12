// Package test ajoute les méthodes utilisées pour les tests
package test

import (
	"testing"
	"time"

	"bou.ke/monkey"
)

// FakeTime méthode qui permet de fausser la méthode `time.Now` en la forçant à toujours retourner
// le paramètre `t`
func FakeTime(test *testing.T, t time.Time) {
	monkey.Patch(time.Now, func() time.Time {
		return t
	})
	test.Cleanup(unfakeTime)
}

func unfakeTime() {
	monkey.Unpatch(time.Now)
}
