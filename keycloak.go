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

type scope []string

func (sc scope) containsRole(role string) bool {
	for _, s := range sc {
		if role == s {
			return true
		}
	}
	return false
}

func (sc scope) containsScope(roles scope) bool {
	for _, r := range roles {
		if !sc.containsRole(r) {
			return false
		}
	}
	return true
}

func (sc scope) overlapsScope(roles scope) bool {
	for _, r := range roles {
		if sc.containsRole(r) {
			return true
		}
	}
	return false
}

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

	_, claims, err := keycloak.DecodeAccessToken(context.Background(), rawToken[1], viper.GetString("keycloakRealm"))

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

	if email, ok := (*claims)["email"]; ok {
		c.Set("userID", email)
	} else {
		c.AbortWithStatus(401)
	}

	c.Set("claims", claims)

	c.Next()
}

func fakeCloakMiddleware(c *gin.Context) {
	var fakeRoles []interface{}
	for _, r := range viper.GetStringSlice("fakeKeycloakRoles") {
		fakeRoles = append(fakeRoles, r)
	}

	var claims = jwt.MapClaims{
		"resource_access": map[string]interface{}{
			viper.GetString("keycloakClient"): map[string]interface{}{
				"roles": fakeRoles,
			},
		},
	}

	c.Set("userID", viper.GetString("fakeUserIDKeycloak"))
	c.Set("claims", &claims)
	c.Next()
}

func scopeFromClaims(claims *jwt.MapClaims) scope {
	var resourceAccess = make(map[string]interface{})
	var client = make(map[string]interface{})
	var scope []interface{}

	resourceAccessInterface, ok := (*claims)["resource_access"]
	if ok {
		resourceAccess, _ = resourceAccessInterface.(map[string]interface{})
	}

	clientInterface, ok := (resourceAccess)[viper.GetString("keycloakClient")]
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

func scopeFromContext(c *gin.Context) scope {
	claims, ok := c.Get("claims")
	if !ok {
		return nil
	}
	roles := scopeFromClaims(claims.(*jwt.MapClaims))
	return roles
}

func (sc scope) zoneGeo() []string {
	var zone []string
	for _, role := range sc {
		departements, _ := ref.zones[role]
		zone = append(zone, departements...)
	}
	for _, s := range sc {
		zone = append(zone, s)
	}
	return zone
}
