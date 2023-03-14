package ops

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/signaux-faibles/datapi/src/core"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"sync"
)

// ConfigureEndpoint configure le endpoint du package `ops`
func ConfigureEndpoint(api *gin.Engine) {
	utils := api.Group("/utils", adminAuthMiddleware())
	utils.GET("/import", importHandler) // 1
	utils.GET("/keycloak", getKeycloakUsers)
	utils.GET("/metrics", gin.WrapH(promhttp.Handler()))
	// utils.GET("/wekanImport", wekanImportHandler)
	utils.GET("/sireneImport", sireneImportHandler)   // 2
	utils.GET("/listImport/:algo", listImportHandler) // 3
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

func sireneImportHandler(c *gin.Context) {
	if viper.GetString("sireneULPath") == "" || viper.GetString("geoSirenePath") == "" {
		c.AbortWithStatusJSON(http.StatusConflict, "not supported, missing parameters in server configuration")
		return
	}
	log.Println("Truncate etablissement & entreprise table")

	err := core.TruncateSirens()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	log.Println("Tables truncated")

	wg := sync.WaitGroup{}
	wg.Add(2)
	ctx, cancelCtx := context.WithCancel(context.Background())
	go core.InsertSireneUL(ctx, cancelCtx, &wg)
	go core.InsertGeoSirene(ctx, cancelCtx, &wg)
	wg.Wait()
}
