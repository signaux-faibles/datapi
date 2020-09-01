package main

import (
	"context"
	"log"
	"strings"

	gocloak "github.com/Nerzal/gocloak/v6"
	"github.com/dgrijalva/jwt-go/v4"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"github.com/spf13/viper"
)

type scope []string

type keycloakUser struct {
	userName  *string
	email     *string
	firstName *string
	lastName  *string
	roles     scope
}

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
		c.Set("username", email)
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

	c.Set("username", viper.GetString("fakeUsernameKeycloak"))
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

func getKeycloakUsers(c *gin.Context) {
	roles := scopeFromContext(c)
	if !roles.containsRole("import") {
		c.AbortWithStatus(403)
		return
	}

	userMap, roleMap, err := fetchUsersAndRoles()
	if err != nil {
		c.JSON(err.Code(), err.Error())
		return
	}
	err = importUsersAndRoles(userMap, roleMap)
	if err != nil {
		c.JSON(err.Code(), err.Error())
	}
	c.JSON(200, "utilisateurs mis à jour")
}

func fetchUsersAndRoles() (map[string]keycloakUser, map[string]*string, Jerror) {
	realm := viper.GetString("keycloakRealm")
	client := viper.GetString("keycloakClient")
	kcAdmin := gocloak.NewClient(viper.GetString("keycloakHostname"))
	jwt, err := kcAdmin.LoginAdmin(
		context.TODO(),
		viper.GetString("keycloakAdminUsername"),
		viper.GetString("keycloakAdminPassword"),
		realm,
	)
	if err != nil {
		return nil, nil, errorToJSON(500, err)
	}

	users, err := kcAdmin.GetUsers(
		context.TODO(),
		jwt.AccessToken,
		viper.GetString("keycloakRealm"),
		gocloak.GetUsersParams{},
	)
	if err != nil {
		return nil, nil, errorToJSON(500, err)
	}

	clients, err := kcAdmin.GetClients(
		context.Background(),
		jwt.AccessToken,
		realm,
		gocloak.GetClientsParams{
			ClientID: &client,
		},
	)
	if err != nil {
		return nil, nil, errorToJSON(500, err)
	}
	if len(clients) != 1 {
		return nil, nil, newJSONerror(500, "client not found")
	}

	userMap := make(map[string]keycloakUser)
	for _, u := range users {
		roles, err := kcAdmin.GetClientRolesByUserID(
			context.TODO(),
			jwt.AccessToken,
			realm,
			*clients[0].ID,
			*u.ID,
		)
		if err != nil {
			return nil, nil, errorToJSON(500, err)
		}
		var userRoles []string
		for _, r := range roles {
			userRoles = append(userRoles, *r.Name)
		}
		userMap[*u.ID] = keycloakUser{
			userName:  u.Username,
			email:     u.Email,
			firstName: u.FirstName,
			lastName:  u.LastName,
			roles:     userRoles,
		}
	}

	if len(userMap) == 0 {
		return nil, nil, newJSONerror(204, "no users")
	}

	roles, err := kcAdmin.GetClientRoles(
		context.Background(),
		jwt.AccessToken,
		realm,
		*clients[0].ID)

	roleMap := make(map[string]*string)
	for _, r := range roles {
		roleMap[*r.Name] = r.Description
	}

	return userMap, roleMap, nil
}

func importUsersAndRoles(userMap map[string]keycloakUser, roleMap map[string]*string) Jerror {
	tx, err := db.Begin(context.Background())
	if err != nil {
		return errorToJSON(500, err)
	}
	defer tx.Commit(context.Background())

	batch := pgx.Batch{}

	batch.Queue("delete from users;")
	batch.Queue("delete from roles;")

	sqlUser := `insert into users (id, userName, firstName, lastName, roles)
		values ($1, $2, $3, $4, $5);`
	for id, user := range userMap {
		batch.Queue(sqlUser, id, user.userName, user.firstName, user.lastName, user.roles)
	}

	sqlRole := `insert into roles (code, description)
		values ($1, $2);`
	for code, description := range roleMap {
		batch.Queue(sqlRole, code, description)
	}

	b := tx.SendBatch(context.Background(), &batch)
	err = b.Close()
	if err != nil {
		return errorToJSON(500, err)
	}

	return nil
}

type user struct {
	Username  *string `json:"email"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
}

func getUser(username string) (user, error) {
	var u user
	err := db.QueryRow(context.Background(),
		`select username, firstName, lastname from users where username = $1`,
		username).Scan(
		&u.Username,
		&u.FirstName,
		&u.LastName,
	)
	return u, err
}
