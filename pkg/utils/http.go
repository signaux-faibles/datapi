package utils

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func ElapsedTimeHandler(c *gin.Context) {
	start := time.Now()

	// Appel de la logique du prochain middleware ou du handler
	c.Next()

	elapsed := time.Since(start)

	fmt.Printf("Appel %s / Temps écoulé: %s\n", c.Request.URL, elapsed)
}

func GetHTTPParameter[T any](c *gin.Context, paramName string, converter func(string) (T, error)) (T, error) {
	var empty T
	param := c.Param(paramName)
	if len(param) == 0 {
		return empty, errors.Errorf("pas de paramètre '%s'", paramName)
	}
	r, err := converter(param)
	if err != nil {
		return empty, errors.Wrap(err, fmt.Sprintf("le paramètre '%s' n'est pas du bon type", paramName))
	}
	return r, nil
}

func GetIntHTTPParameter(c *gin.Context, paramName string) (int, error) {
	return GetHTTPParameter(c, paramName, strconv.Atoi)
}

func GetDateHTTPParameter(c *gin.Context, paramName string, format string) (time.Time, error) {
	return GetHTTPParameter(c, paramName, func(s string) (time.Time, error) { return time.Parse(format, s) })
}
