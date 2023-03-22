// Package ops contient le code lié aux opérations d'administration dans datapi
package ops

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
	endpoint.GET("/import", importHandler) // 1
	endpoint.GET("/keycloak", keycloakUsersHandler)
	endpoint.GET("/metrics", gin.WrapH(promhttp.Handler()))
	endpoint.GET("/sireneImport", importSireneHandler)       // 2
	endpoint.GET("/importListes/:algo", importListesHandler) // 3
	endpoint.GET("/mep", importHandler, importSireneHandler)
	endpoint.GET("/mep/:algo", importHandler, importSireneHandler, importListesHandler)
}

func importSireneHandler(c *gin.Context) {
	err := importSirene()
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	c.Next()
	//c.JSON(http.StatusOK, "sirenes mis à jour")
}

func importHandler(c *gin.Context) {
	err := importEntreprisesAndEtablissement()
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	c.Next()
	//c.JSON(http.StatusOK, "entreprises & etablissements mis à jour")
}

func importListesHandler(c *gin.Context) {
	algo := c.Params.ByName("algo")
	err := importListes(algo)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	c.Next()
	//c.JSON(http.StatusOK, "entreprises & etablissements mis à jour")
}

func keycloakUsersHandler(c *gin.Context) {
	err := getKeycloakUsers()
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "utilisateurs mis à jour"})
}
