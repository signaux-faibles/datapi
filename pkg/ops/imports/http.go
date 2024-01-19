// Package imports contient le code lié aux opérations d'administration dans datapi
package imports

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"datapi/pkg/utils"
)

// ConfigureEndpoint configure l'endpoint du package `ops`
func ConfigureEndpoint(endpoint *gin.RouterGroup) {
	endpoint.GET("/sirene/stocketablissement", importStockEtablissementsHandler)
	endpoint.GET("/sirene/unitelegale", importUnitesLegalesHandler)
	endpoint.GET("/liste/:batchNumber/:algo", importListesHandler)
	endpoint.GET("/bce", importBCEHandler)
	endpoint.GET("/paydex", importPaydexHandler)
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

func importListesHandler(c *gin.Context) {
	algo := c.Params.ByName("algo")
	batchNumber := c.Params.ByName("batchNumber")
	if algo == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"erreur": "le paramètre `algo` est obligatoire"})
		return
	}

	err := importListe(c, batchNumber, algo)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
}

func importPaydexHandler(c *gin.Context) {
	err := importPaydex(c)
	if err != nil {
		handlerError := c.AbortWithError(http.StatusInternalServerError, err)
		if handlerError != nil {
			slog.Error("erreur lors de l'arrêt du handler paydex", slog.Any("error", err))
		}
	}
}
