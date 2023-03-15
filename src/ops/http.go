package ops

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/signaux-faibles/datapi/src/utils"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

// ConfigureEndpoint configure l'endpoint du package `ops`
func ConfigureEndpoint(path string, api *gin.Engine) {
	endpoint := api.Group(path, adminAuthMiddleware())
	endpoint.GET("/import", importHandler) // 1
	endpoint.GET("/keycloak", keycloakUsersHandler)
	endpoint.GET("/metrics", gin.WrapH(promhttp.Handler()))
	endpoint.GET("/sireneImport", importSireneHandler)       // 2
	endpoint.GET("/importListes/:algo", importListesHandler) // 3
	endpoint.GET("/mep", importHandler, importSireneHandler)
	endpoint.GET("/mep/:algo", importHandler, importSireneHandler, importListesHandler)
}

func adminAuthMiddleware() gin.HandlerFunc {
	var whitelist = viper.GetStringSlice("adminWhitelist")
	var wlmap = make(map[string]bool)
	for _, ip := range whitelist {
		wlmap[ip] = true
	}

	return func(c *gin.Context) {
		if !wlmap[c.ClientIP()] {
			log.Printf("Connection from %s is not granted in adminWhitelist, see config.toml\n", c.ClientIP())
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
	}
}

func importSireneHandler(c *gin.Context) {
	err := importSirene()
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, "sirenes mis à jour")
}

func importHandler(c *gin.Context) {
	err := importEntreprisesAndEtablissement()
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, "entreprises & etablissements mis à jour")
}

func importListesHandler(c *gin.Context) {
	algo := c.Params.ByName("algo")
	err := importListes(algo)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, "entreprises & etablissements mis à jour")
}

func keycloakUsersHandler(c *gin.Context) {
	err := getKeycloakUsers()
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, "utilisateurs mis à jour")
}
