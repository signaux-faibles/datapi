package main

import (
	"time"

	"github.com/gin-gonic/gin"
	dalib "github.com/signaux-faibles/datapi/lib"

	jwt "github.com/appleboy/gin-jwt"
	"github.com/spf13/viper"
)

var authMiddleware, err = jwt.New(&jwt.GinJWTMiddleware{
	Realm:           "Signaux-Faibles",
	Key:             []byte(viper.GetString("jwtSecret")),
	SendCookie:      false,
	Timeout:         time.Hour,
	MaxRefresh:      time.Hour,
	IdentityKey:     "id",
	PayloadFunc:     payloadHandler,
	IdentityHandler: identityHandler,
	Authenticator:   authenticatorHandler,
	Authorizator:    authorizatorHandler,
	Unauthorized:    unauthorizedHandler,
	TokenLookup:     "header: Authorization, query: token",
	TokenHeadName:   "Bearer",
	TimeFunc:        time.Now,
})

func payloadHandler(data interface{}) jwt.MapClaims {
	if v, ok := data.(dalib.User); ok {
		return jwt.MapClaims{
			"email": v.Email,
			"scope": v.Scope,
		}
	}
	return jwt.MapClaims{}
}

func identityHandler(c *gin.Context) interface{} {
	claims := jwt.ExtractClaims(c)
	var scope []interface{}
	if claims["scope"] != nil {
		scope = claims["scope"].([]interface{})
	}
	email := claims["email"].(string)
	user := dalib.User{
		Email: email,
		Scope: dalib.ToScope(scope),
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
	// if credentials.Email == "admin" && credentials.Password == "admin" {
	// 	return dalib.User{
	// 		Email: "admin@test.com",
	// 		Scope: dalib.Tags{"admin"},
	// 	}, nil
	// }

	// if credentials.Email == "user" && credentials.Password == "user" {
	// 	return dalib.User{
	// 		Email: "user@test.com",
	// 		Scope: dalib.Tags{"user"},
	// 	}, nil
	// }
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
