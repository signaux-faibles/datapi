package ops

import (
	"context"
	"github.com/Nerzal/gocloak/v10"
	"github.com/jackc/pgx/v4"
	"github.com/signaux-faibles/datapi/src/core"
	"github.com/signaux-faibles/datapi/src/db"
	"github.com/signaux-faibles/datapi/src/utils"
	"github.com/spf13/viper"
	"net/http"
)

type keycloakUser struct {
	UserName  *string
	Email     *string
	FirstName *string
	LastName  *string
	Roles     core.Scope
}

func getKeycloakUsers() error {
	var err error
	userMap, roleMap, err := fetchUsersAndRoles()
	if err != nil {
		return err
	}
	err = importUsersAndRoles(userMap, roleMap)
	if err != nil {
		return err
	}
	return nil
}

func fetchUsersAndRoles() (map[string]keycloakUser, map[string]*string, error) {
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
		return nil, nil, err
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
		return nil, nil, err
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
		return nil, nil, err
	}
	if len(clients) != 1 {
		return nil, nil, err
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
			return nil, nil, err
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
		return nil, nil, utils.NewJSONerror(http.StatusNoContent, "no users")
	}

	roles, err := kcAdmin.GetClientRoles(
		context.Background(),
		jwt.AccessToken,
		realm,
		*clients[0].ID,
		gocloak.GetRoleParams{})

	if err != nil {
		return nil, nil, err
	}

	roleMap := make(map[string]*string)
	for _, r := range roles {
		roleMap[*r.Name] = r.Description
	}

	return userMap, roleMap, nil
}

func importUsersAndRoles(userMap map[string]keycloakUser, roleMap map[string]*string) error {
	tx, err := db.Get().Begin(context.Background())
	if err != nil {
		return err
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
		return err
	}

	return nil
}
