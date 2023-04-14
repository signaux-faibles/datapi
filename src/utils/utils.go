// Package utils contient le code technique commun
package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
	"net/http"
)

// AbortWithError gère le statut http en fonction de l'erreur
func AbortWithError(c *gin.Context, err error) {
	message := err.Error()
	code := http.StatusInternalServerError
	if err, ok := err.(Jerror); ok {
		code = err.Code()
	}
	c.AbortWithStatusJSON(code, message)
}

// Convert applique la fonction `transformer` à tous les éléments d'un slice et retourne le tableau convertit
func Convert[I interface{}, O interface{}](values []I, transformer func(I) O) []O {
	var output []O
	for _, current := range values {
		output = append(output, transformer(current))
	}
	return output
}

// Contains retourne `vrai` si array contient `test`
func Contains[T interface{}](values []T, searched T) bool {
	return ContainsOnConditions(
		values,
		searched,
		func(t1 T, t2 T) bool { return cmp.Equal(t1, t2) },
	)
}

// ContainsOnConditions retourne `vrai` si array contient `test`
func ContainsOnConditions[T interface{}](values []T, searched T, conditions ...func(t1 T, t2 T) bool) bool {
	for _, current := range values {
		accept := true
		for _, condition := range conditions {
			accept = accept && condition(current, searched)
		}
		if accept {
			return true
		}
	}
	return false
}

func GetKeys[K comparable, V any](input map[K]V) []K {
	keys := make([]K, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	return keys
}

// ToMap crée une map d'après une liste de valeur
func ToMap[K comparable, V interface{}](values []V, transformer func(V) K) map[K]V {
	r := make(map[K]V)
	for _, current := range values {
		key := transformer(current)
		r[key] = current
	}
	return r
}

func Coalesce[T any](pointers ...*T) *T {
	for _, i := range pointers {
		if i != nil {
			return i
		}
	}
	return nil
}
