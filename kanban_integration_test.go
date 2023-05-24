//go:build integration

package main

import (
	"github.com/jaswdr/faker"
	"github.com/signaux-faibles/datapi/src/test"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

var fake faker.Faker

func init() {
	fake = faker.New()
}

func TestApi_Wekan_endpoint(t *testing.T) {
	someRoles := []string{fake.App().Name(), fake.App().Name(), fake.App().Name()}
	someRolesWithWekan := append(someRoles, "wekan")

	type args struct {
		roles []string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"l'utilisateur n'a pas le rôle `wekan`", args{someRoles}, http.StatusForbidden},
		{"l'utilisateur a le rôle `wekan`", args{someRolesWithWekan}, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ass := assert.New(t)
			changeRoles(t, tt.args.roles...)
			response := test.HTTPGet(t, "/kanban/config")
			ass.Equal(tt.want, response.StatusCode)
		})
	}

}

func changeRoles(t *testing.T, roles ...string) {
	// give kanban role
	rolesBackup := viper.Get("fakeKeycloakRoles")
	t.Cleanup(func() {
		t.Logf("rollback des roles après le test '%s' : %s", t.Name(), rolesBackup)
		viper.Set("fakeKeycloakRoles", rolesBackup)
	})
	viper.Set("fakeKeycloakRoles", roles)
}
