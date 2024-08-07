package core

import (
	"context"
	"datapi/pkg/db"
	"datapi/pkg/utils"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/Nerzal/gocloak/v10"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
)

// Scope représente les habilitations d'un utilisateur (rôle vs département)
type Scope []string

func connectKC() gocloak.GoCloak {
	keycloak := gocloak.NewClient(viper.GetString("keycloakHostname"))
	if keycloak == nil {
		panic("keycloak not available")
	}
	return keycloak
}

func getRawToken(c *gin.Context) ([]string, error) {
	if len(c.Request.Header["Authorization"]) == 0 {
		return nil, errors.New("no authorization header")
	}
	header := c.Request.Header["Authorization"][0]
	rawToken := strings.Split(header, " ")
	if len(rawToken) != 2 {
		return nil, errors.New("erreur sur le header 'Authorization' header")
	}
	return rawToken, nil
}

func keycloakMiddleware(c *gin.Context) {
	rawToken, err := getRawToken(c)
	if err != nil {
		log.Println(err.Error())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	_, claims, err := keycloak.DecodeAccessToken(context.Background(), rawToken[1], viper.GetString("keycloakRealm"))

	if err != nil {
		log.Println("unable to decode token: " + err.Error())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if errValid := claims.Valid(); err != nil && errValid != nil {
		log.Println("token is invalid: " + errValid.Error())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if username, ok := (*claims)["preferred_username"]; ok {
		c.Set("username", username)
	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, "username absent")
	}

	if givenName, ok := (*claims)["given_name"]; ok {
		c.Set("given_name", givenName)
	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, "given_name absent")
	}

	if givenName, ok := (*claims)["family_name"]; ok {
		c.Set("family_name", givenName)
	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, "family_name absent")
	}

	c.Set("claims", claims)

	c.Next()
}

func fakeCloakMiddleware(c *gin.Context) {
	var fakeRoles []string
	fakeRoles = append(fakeRoles, viper.GetStringSlice("fakeKeycloakRoles")...)
	middleware := addFakeRolesMiddleware(fakeRoles...)
	middleware(c)
}

// addFakeRolesMiddleware ajoute les roles passés en paramètre dans le contexte `gin`
// comme s'il s'agissait de roles venant de keycloak
func addFakeRolesMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var claims = jwt.MapClaims{
			"resource_access": map[string]interface{}{
				viper.GetString("keycloakClient"): map[string]interface{}{
					"roles": toInterfaces(roles),
				},
			},
		}

		c.Set("username", viper.GetString("fakeUsernameKeycloak"))
		c.Set("given_name", "John")
		c.Set("family_name", "Doe")
		c.Set("claims", &claims)
	}
}

func scopeFromClaims(claims *jwt.MapClaims) Scope {
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

func scopeFromContext(c *gin.Context) Scope {
	claims, ok := c.Get("claims")
	if !ok {
		return nil
	}
	roles := scopeFromClaims(claims.(*jwt.MapClaims))
	return roles
}

type keycloakUser struct {
	Username  *string `json:"username"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
}

func getUser(username string) (keycloakUser, error) {
	var u keycloakUser
	err := db.Get().QueryRow(context.Background(),
		`select username, firstName, lastname from users where username = $1`,
		username).Scan(
		&u.Username,
		&u.FirstName,
		&u.LastName,
	)
	return u, err
}

func toInterfaces[I any](items []I) []interface{} {
	return utils.Convert(items, func(i I) interface{} { return interface{}(i) })
}
