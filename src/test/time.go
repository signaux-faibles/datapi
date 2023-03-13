// Package test ajoute les méthodes utilisées pour les tests
package test

import (
	"bou.ke/monkey"
	"time"
)

// FakeTime méthode qui permet de fausser la méthode `time.Now` en la forçant à toujours retourner
// le paramètre `t`
func FakeTime(t time.Time) {
	monkey.Patch(time.Now, func() time.Time {
		return t
	})
}
