// Package misc contient le code des handler http divers
package misc

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/signaux-faibles/datapi/src/core"
	"github.com/signaux-faibles/datapi/src/utils"
	"net/http"
)

// ConfigureEndpoint configure l'endpoint du package `ops`
func ConfigureEndpoint(path string, api *gin.Engine) {
	endpoint := api.Group(path, core.AdminAuthMiddleware)
	endpoint.GET("/keycloak", keycloakUsersHandler)
	endpoint.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

func keycloakUsersHandler(c *gin.Context) {
	err := getKeycloakUsers()
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "utilisateurs mis à jour"})
}
