package main

import (
	"strings"

	gocloak "github.com/Nerzal/gocloak/v5"
	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func connectKC() *gocloak.GoCloak {
	keycloak := gocloak.NewClient(viper.GetString("keycloakHostname"))
	return &keycloak
}

func keycloakMiddleware(c *gin.Context) {
	header := c.Request.Header["Authorization"][0]
	rawToken := strings.Split(header, " ")[1]

	token, claims, err := keycloak.DecodeAccessToken(rawToken, viper.GetString("keycloakRealm"))
	if errValid := claims.Valid(); err != nil && errValid != nil {
		c.AbortWithStatus(401)
	}

	c.Set("token", token)
	c.Set("claims", claims)

	c.Next()
}

func scopeFromClaims(claims *jwt.MapClaims) []string {
	var resourceAccess = make(map[string]interface{})
	var client = make(map[string]interface{})
	var scope []interface{}

	resourceAccessInterface, ok := (*claims)["resource_access"]
	if ok {
		resourceAccess, _ = resourceAccessInterface.(map[string]interface{})
	}

	clientInterface, ok := (resourceAccess)["signauxfaibles"]
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
