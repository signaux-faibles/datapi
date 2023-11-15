// Package imports contient le code lié aux opérations d'administration dans datapi
package imports

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"datapi/pkg/utils"
)

// ConfigureEndpoint configure l'endpoint du package `ops`
func ConfigureEndpoint(endpoint *gin.RouterGroup) {
	endpoint.GET("/ee", importEtablissementHandler)
	endpoint.GET("/sirene/stocketablissement", importStockEtablissementsHandler)
	endpoint.GET("/sirene/unitelegale", importUnitesLegalesHandler)
	endpoint.GET("/listes/:algo", importListesHandler)
	endpoint.GET("/full", importEtablissementHandler, importStockEtablissementsHandler, importUnitesLegalesHandler)
	endpoint.GET("/full/:algo", importEtablissementHandler, importStockEtablissementsHandler, importUnitesLegalesHandler, importListesHandler)
	endpoint.GET("/bce", importBCEHandler)
	endpoint.GET("/paydexhisto", importPaydexHistoHandler)
	endpoint.GET("/urssaf", importUrssafHandler)
}

func importStockEtablissementsHandler(c *gin.Context) {
	err := importStockEtablissement(c)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
}

func importUnitesLegalesHandler(c *gin.Context) {
	err := importUnitesLegales(c)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
}

func importEtablissementHandler(c *gin.Context) {
	err := importEtablissement()
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
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
}

func importPaydexHistoHandler(c *gin.Context) {
	err := importPaydexHisto(c)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
}
