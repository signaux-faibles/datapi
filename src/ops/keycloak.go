package ops

import (
	"context"
	"github.com/Nerzal/gocloak/v10"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"github.com/signaux-faibles/datapi/src/core"
	"github.com/signaux-faibles/datapi/src/db"
	"github.com/signaux-faibles/datapi/src/utils"
	"github.com/spf13/viper"
	"log"
)

type keycloakUser struct {
	UserName  *string
	Email     *string
	FirstName *string
	LastName  *string
	Roles     core.Scope
}

func getKeycloakUsers(c *gin.Context) {
	userMap, roleMap, err := fetchUsersAndRoles()
	if err != nil {
		c.JSON(err.Code(), err.Error())
		return
	}
	err = importUsersAndRoles(userMap, roleMap)
	if err != nil {
		log.Print(err.Error())
		c.JSON(err.Code(), err.Error())
		return
	}
	c.JSON(200, "utilisateurs mis à jour")
}

func fetchUsersAndRoles() (map[string]keycloakUser, map[string]*string, utils.Jerror) {
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
		return nil, nil, utils.ErrorToJSON(500, err)
	}

	var i = 100000
	users, err := kcAdmin.GetUsers(
		context.TODO(),
		jwt.AccessToken,
		viper.GetString("keycloakRealm"),
		gocloak.GetUsersParams{
			Max: &i,
		},
	)
	if err != nil {
		return nil, nil, utils.ErrorToJSON(500, err)
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
		return nil, nil, utils.ErrorToJSON(500, err)
	}
	if len(clients) != 1 {
		return nil, nil, utils.NewJSONerror(500, "client not found")
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
			return nil, nil, utils.ErrorToJSON(500, err)
		}
		var userRoles []string
		for _, r := range roles {
			userRoles = append(userRoles, *r.Name)
		}
		userMap[*u.ID] = keycloakUser{
			UserName:  u.Username,
			Email:     u.Email,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Roles:     userRoles,
		}
	}

	if len(userMap) == 0 {
		return nil, nil, utils.NewJSONerror(204, "no users")
	}

	roles, err := kcAdmin.GetClientRoles(
		context.Background(),
		jwt.AccessToken,
		realm,
		*clients[0].ID,
		gocloak.GetRoleParams{})

	if err != nil {
		return nil, nil, utils.ErrorToJSON(500, err)
	}

	roleMap := make(map[string]*string)
	for _, r := range roles {
		roleMap[*r.Name] = r.Description
	}

	return userMap, roleMap, nil
}

func importUsersAndRoles(userMap map[string]keycloakUser, roleMap map[string]*string) utils.Jerror {
	tx, err := db.Get().Begin(context.Background())
	if err != nil {
		return utils.ErrorToJSON(500, err)
	}
	defer tx.Commit(context.Background())

	batch := pgx.Batch{}

	batch.Queue("delete from users;")
	batch.Queue("delete from roles;")

	sqlUser := `insert into users (id, userName, firstName, lastName, roles)
		values ($1, $2, $3, $4, $5);`
	for id, user := range userMap {
		batch.Queue(sqlUser, id, user.UserName, user.FirstName, user.LastName, user.Roles)
	}

	sqlRole := `insert into roles (code, description)
		values ($1, $2);`
	for code, description := range roleMap {
		batch.Queue(sqlRole, code, description)
	}

	b := tx.SendBatch(context.Background(), &batch)
	err = b.Close()
	if err != nil {
		return utils.ErrorToJSON(500, err)
	}

	return nil
}
