// Package imports contient le code lié aux opérations d'administration dans datapi
package imports

import (
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/datapi/src/core"
	"github.com/signaux-faibles/datapi/src/utils"
	"net/http"
)

// ConfigureEndpoint configure l'endpoint du package `ops`
func ConfigureEndpoint(path string, api *gin.Engine) {
	endpoint := api.Group(path, core.AdminAuthMiddleware)
	endpoint.GET("/ee", importEntrepriseAndEtablissementHandler) // 1
	endpoint.GET("/sirene", importSireneHandler)                 // 2
	endpoint.GET("/listes/:algo", importListesHandler)           // 3
	endpoint.GET("/full", importEntrepriseAndEtablissementHandler, importSireneHandler)
	endpoint.GET("/full/:algo", importEntrepriseAndEtablissementHandler, importSireneHandler, importListesHandler)
}

func importSireneHandler(c *gin.Context) {
	err := importSirene()
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	//c.Next()
	//c.JSON(http.StatusOK, "sirenes mis à jour")
}

func importEntrepriseAndEtablissementHandler(c *gin.Context) {
	err := importEntreprisesAndEtablissement()
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	//c.Next()
	//c.JSON(http.StatusOK, "entreprises & etablissements mis à jour")
}

func importListesHandler(c *gin.Context) {
	algo := c.Params.ByName("algo")
	if algo == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"erreur": "le paramètre `algo` est obligatoire"})
		return
	}
	err := importListes(algo)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	//c.Next()
	//c.JSON(http.StatusOK, "entreprises & etablissements mis à jour")
}
