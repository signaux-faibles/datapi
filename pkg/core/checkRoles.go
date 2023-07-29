package core

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func CheckAllRolesMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var s Session
		s.Bind(c)
		for _, role := range roles {
			if !s.hasRole(role) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": s.Username + " n'a pas le rôle " + role})
				return
			}
		}
	}
}

func CheckAnyRolesMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(roles) == 0 {
			return
		}
		var s Session
		s.Bind(c)
		for _, role := range roles {
			if s.hasRole(role) {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(
			http.StatusForbidden,
			gin.H{"message": s.Username + " n'a aucun rôle suffisant : " + strings.Join(roles, ", ")},
		)
	}
}
