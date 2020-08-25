package main

import (
	"context"
	"log"
	"strings"

	gocloak "github.com/Nerzal/gocloak/v6"
	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func connectKC() gocloak.GoCloak {
	keycloak := gocloak.NewClient(viper.GetString("keycloakHostname"))
	if keycloak == nil {
		panic("keycloak not available")
	}
	return keycloak
}

func keycloakMiddleware(c *gin.Context) {
	if len(c.Request.Header["Authorization"]) == 0 {
		log.Println("no authorization header")
		c.AbortWithStatus(401)
		return
	}
	header := c.Request.Header["Authorization"][0]

	rawToken := strings.Split(header, " ")
	if len(rawToken) != 2 {
		log.Println("malformed authorization header")
		c.AbortWithStatus(401)
		return
	}

	token, claims, err := keycloak.DecodeAccessToken(context.Background(), rawToken[1], viper.GetString("keycloakRealm"))

	if err != nil {
		log.Println("unable to decode token: " + err.Error())
		c.AbortWithStatus(401)
		return
	}

	if errValid := claims.Valid(); err != nil && errValid != nil {
		log.Println("token is invalid: " + errValid.Error())
		c.AbortWithStatus(401)
		return
	}

	c.Set("token", token)
	c.Set("claims", claims)

	c.Next()
}

func fakeCloakMiddleware(c *gin.Context) {
	var claims jwt.MapClaims
	c.Set("claims", claims)
}

func scopeFromClaims(claims *jwt.MapClaims) []string {
	var resourceAccess = make(map[string]interface{})
	var client = make(map[string]interface{})
	var scope []interface{}

	resourceAccessInterface, ok := (*claims)["resource_access"]
	if ok {
		resourceAccess, _ = resourceAccessInterface.(map[string]interface{})
	}

	clientInterface, ok := (resourceAccess)[viper.GetString("")]
	if ok {
		client, _ = clientInterface.(map[string]interface{})
	}

	scopeInterface, ok := (client)["roles"]
	if ok {
		scope, _ = scopeInterface.([]interface{})
	}

	var tags []string
	for _, tag := range scope {
		tags = append(tags, tag.(string))
	}

	return tags
}
