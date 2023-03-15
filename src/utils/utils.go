// Package utils contient le code technique commun
package utils

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Contains retourne `vrai` si array contient `test`
func Contains(array []string, test string) bool {
	for _, s := range array {
		if s == test {
			return true
		}
	}
	return false
}

// AbortWithError gère le statut http en fonction de l'erreur
func AbortWithError(c *gin.Context, err error) {
	message := err.Error()
	code := http.StatusInternalServerError
	if err, ok := err.(Jerror); ok {
		code = err.Code()
	}
	c.AbortWithStatusJSON(code, message)
}
