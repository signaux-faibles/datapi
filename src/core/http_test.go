package core

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func init() {
	viper.Set("keycloakClient", "signauxfaibles")
}

func TestCheckAllRolesMiddleware(t *testing.T) {
	ass := assert.New(t)

	username := fake.Person().Name()

	someRoles := generateSomeRoles()
	otherRoles := generateSomeRoles()
	sameRolesAndMore := append(someRoles, otherRoles...)

	type args struct {
		neededRoles   []string
		providedRoles []string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"tous les rôles nécessaires sont présents", args{neededRoles: someRoles, providedRoles: sameRolesAndMore}, http.StatusOK},
		{"il manque au moins 1 rôle", args{neededRoles: sameRolesAndMore, providedRoles: someRoles}, http.StatusForbidden},
		{"aucun rôle n'est nécessaire", args{neededRoles: nil, providedRoles: someRoles}, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			_, routeur := gin.CreateTestContext(recorder)
			routeur.Handle(
				"GET",
				"/test/any_roles",
				addFakeRolesMiddleware(username, tt.args.providedRoles...),
				CheckAllRolesMiddleware(tt.args.neededRoles...),
				ok,
			)

			request, err := http.NewRequest(http.MethodGet, "/test/any_roles", nil)
			assert.NoError(t, err)

			routeur.ServeHTTP(recorder, request)
			t.Logf("message de réponse : %d - %s", recorder.Code, recorder.Body.Bytes())
			ass.Equal(tt.want, recorder.Code)
		})
	}
}

func TestCheckAnyRolesMiddleware(t *testing.T) {
	ass := assert.New(t)

	username := fake.Person().Name()

	aPreciseRole := fake.Color().ColorName()

	someRoles := append(generateSomeRoles(), aPreciseRole)
	otherRoles := generateSomeRoles()
	otherRolesWithPreciseRole := append(otherRoles, aPreciseRole)

	type args struct {
		neededRoles   []string
		providedRoles []string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"l'utilisateur a au moins 1 des rôles", args{neededRoles: someRoles, providedRoles: otherRolesWithPreciseRole}, http.StatusOK},
		{"l'utilisateur n'a aucun des rôles demandés", args{neededRoles: someRoles, providedRoles: otherRoles}, http.StatusForbidden},
		{"l'utilisateur n'a aucun role", args{neededRoles: someRoles, providedRoles: nil}, http.StatusForbidden},
		{"aucun rôle n'est de demandé", args{neededRoles: nil, providedRoles: someRoles}, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			_, routeur := gin.CreateTestContext(recorder)
			routeur.Handle(
				"GET",
				"/test/any_roles",
				addFakeRolesMiddleware(username, tt.args.providedRoles...),
				CheckAnyRolesMiddleware(tt.args.neededRoles...),
				ok,
			)

			request, err := http.NewRequest(http.MethodGet, "/test/any_roles", nil)
			assert.NoError(t, err)

			routeur.ServeHTTP(recorder, request)
			t.Logf("message de réponse : %d - %s", recorder.Code, recorder.Body.Bytes())
			ass.Equal(tt.want, recorder.Code)
			recorder.Flush()
		})
	}
}

func generateSomeRoles() []string {
	length := fake.IntBetween(1, 3)
	r := make([]string, length)
	for i := 0; i < length; i++ {
		r[i] = fake.Lorem().Word()
	}
	return r
}

func ok(c *gin.Context) {
	c.JSONP(http.StatusOK, gin.H{"message": "c'est encore un succès"})
}
