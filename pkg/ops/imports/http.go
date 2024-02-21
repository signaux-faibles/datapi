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
	endpoint.GET("/ee", importEtablissementHandler)
	endpoint.GET("/sirene/stocketablissement", importStockEtablissementsHandler)
	endpoint.GET("/sirene/unitelegale", importUnitesLegalesHandler)
	endpoint.GET("/liste/:batchNumber/:algo", importPredictionsHandler)
	endpoint.DELETE("/liste/:batchNumber/:algo", deletePredictionsHandler)
	endpoint.GET("/liste/refresh", refreshVtablesHandler)
	endpoint.GET("/full", importEtablissementHandler, importStockEtablissementsHandler, importUnitesLegalesHandler)
	endpoint.GET("/full/:algo", importEtablissementHandler, importStockEtablissementsHandler, importUnitesLegalesHandler, importPredictionsHandler)
	endpoint.GET("/bce", importBCEHandler)
	endpoint.GET("/paydex", importPaydexHandler)
	endpoint.GET("/urssaf", importUrssafHandler)
	endpoint.GET("/urssaf/aggregate", aggregateUrssafTempDataHandler)
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

func importPredictionsHandler(c *gin.Context) {
	algo := c.Params.ByName("algo")
	batchNumber := c.Params.ByName("batchNumber")
	if algo == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"erreur": "le paramètre `algo` est obligatoire"})
		return
	}
	err := importPredictions(batchNumber, algo)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
}

func deletePredictionsHandler(c *gin.Context) {
	algo := c.Params.ByName("algo")
	batchNumber := c.Params.ByName("batchNumber")
	if algo == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"erreur": "le paramètre `algo` est obligatoire"})
		return
	}
	deletePredictionLogger := slog.Default().With(slog.String("algo", algo), slog.String("batch", batchNumber))
	deletePredictionLogger.Info("supprime les prédictions", slog.String("status", "START"))
	_, err := deletePredictions(batchNumber, algo)
	if err != nil {
		slog.Error("erreur pendant la suppression des prédictions", slog.Any("error", err))
		utils.AbortWithError(c, err)
		return
	}
	deletePredictionLogger.Info("supprime les prédictions", slog.String("status", "END"))
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
