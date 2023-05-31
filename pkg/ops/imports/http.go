// Package imports contient le code lié aux opérations d'administration dans datapi
package imports

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"datapi/pkg/utils"
)

// ConfigureEndpoint configure l'endpoint du package `ops`
func ConfigureEndpoint(endpoint *gin.RouterGroup) {
	endpoint.GET("/ee", importEntrepriseAndEtablissementHandler) // 1
	endpoint.GET("/sirene", importSireneHandler)                 // 2
	endpoint.GET("/listes/:algo", importListesHandler)           // 3
	endpoint.GET("/full", importEntrepriseAndEtablissementHandler, importSireneHandler)
	endpoint.GET("/full/:algo", importEntrepriseAndEtablissementHandler, importSireneHandler, importListesHandler)
	endpoint.GET("/bce", importBCEHandler)
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
