package main

import (
	"github.com/gin-gonic/gin"
	dalib "github.com/signaux-faibles/datapi/lib"

	jwt "github.com/appleboy/gin-jwt"
)

func payloadHandler(data interface{}) jwt.MapClaims {
	if v, ok := data.(dalib.User); ok {
		return jwt.MapClaims{
			"email": v.Email,
			"scope": v.Scope,
			"value": v.Value,
		}
	}
	return jwt.MapClaims{}
}

func identityHandler(c *gin.Context) interface{} {
	claims := jwt.ExtractClaims(c)
	email, ok := claims["email"].(string)
	if !ok {
		return nil
	}

	scope, err := dalib.ToScope(claims["scope"])
	if err != nil {
		return nil
	}

	user := dalib.User{
		Email: email,
		Scope: scope,
	}

	return &user
}

func authenticatorHandler(c *gin.Context) (interface{}, error) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBind(&credentials); err != nil {
		return "", jwt.ErrMissingLoginValues
	}

	user, err := dalib.Login(credentials.Email, credentials.Password)

	if err != nil {
		return dalib.User{}, jwt.ErrFailedAuthentication
	}

	return user, nil
}

func authorizatorHandler(data interface{}, c *gin.Context) bool {
	return true
}

func unauthorizedHandler(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"code":    code,
		"message": message,
	})
}
