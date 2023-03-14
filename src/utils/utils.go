package utils

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Contains(array []string, test string) bool {
	for _, s := range array {
		if s == test {
			return true
		}
	}
	return false
}

func AbortWithError(c *gin.Context, err error) {
	if err, ok := err.(Jerror); ok {
		c.AbortWithStatusJSON(err.Code(), err.Error())
		return
	}
	c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
	return
}
