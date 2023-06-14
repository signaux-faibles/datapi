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

// Contains retourne `vrai` si le slice `values` contient l'élément `searched`
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

func Filter[T any](input []T, test func(T) bool) (output []T) {
	for _, element := range input {
		if test(element) {
			output = append(output, element)
		}
	}
	return output
}

func GetKeys[K comparable, V any](input map[K]V) []K {
	keys := make([]K, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	return keys
}

func GetValues[K comparable, V any](input map[K]V) []V {
	keys := make([]V, 0, len(input))
	for _, value := range input {
		keys = append(keys, value)
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

func First[T any](ss []T, test func(T) bool) (ret T, ok bool) {
	for _, s := range ss {
		if test(s) {
			return s, true
		}
	}
	return ret, false
}

func Any[T any](input []T, test func(T) bool) bool {
	for _, element := range input {
		if test(element) {
			return true
		}
	}
	return false
}

func Uniq[Element comparable](array []Element) []Element {
	m := make(map[Element]struct{})
	for _, element := range array {
		m[element] = struct{}{}
	}
	var set []Element
	for element := range m {
		set = append(set, element)
	}
	return set
}

func Overlaps[T comparable](array1 []T, array2 []T) bool {
	for _, t1 := range array1 {
		for _, t2 := range array2 {
			if t1 == t2 {
				return true
			}
		}
	}
	return false
}
