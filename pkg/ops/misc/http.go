// Package misc contient le code des handler http divers
package misc

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"datapi/pkg/utils"
)

// ConfigureEndpoint configure l'endpoint du package `ops`
func ConfigureEndpoint(endpoint *gin.RouterGroup) {
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
