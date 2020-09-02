package main

import (
	"fmt"
	"log"

	gocloak "github.com/Nerzal/gocloak/v6"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/spf13/viper"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var keycloak gocloak.GoCloak
var db *pgxpool.Pool
var ref reference

func main() {
	loadConfig()
	db = connectDB()
	keycloak = connectKC()
	runAPI()
}

func loadConfig() {
	viper.SetDefault("MigrationsDir", "./migrations")
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.ReadInConfig()
}

func runAPI() {
	if viper.GetBool("prod") {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AddAllowHeaders("Authorization")
	config.AddAllowMethods("GET", "POST", "DELETE")
	router.Use(cors.New(config))

	var kcm func(c *gin.Context)
	if viper.GetBool("enableKeycloak") {
		kcm = keycloakMiddleware
	} else {
		kcm = fakeCloakMiddleware
	}

	entreprise := router.Group("/entreprise", kcm)
	entreprise.GET("/get/:siren", validSiren, getEntreprise)
	entreprise.GET("/all/:siren", validSiren, getEntrepriseEtablissements)

	etablissement := router.Group("/etablissement", kcm)
	etablissement.GET("/get/:siret", validSiret, getEtablissement)
	etablissement.GET("/etablissement/comments/:siret", validSiret, getEntrepriseComments)
	etablissement.POST("/etablissement/comments/:siret", validSiret, addEntrepriseComment)
	etablissement.PUT("/etablissement/comments/:id", updateEntrepriseComment)

	follow := router.Group("/follow", kcm)
	follow.GET("", getEtablissementsFollowedByCurrentUser)
	follow.POST("/:siret", validSiret, followEtablissement)
	follow.DELETE("/:siret", validSiret, unfollowEtablissement)

	listes := router.Group("/listes", kcm)
	listes.GET("", getListes)

	scores := router.Group("/scores", kcm)
	scores.POST("", getLastListeScores)
	scores.POST("/:id", getListeScores)

	reference := router.Group("/reference", kcm)
	reference.GET("/naf", getCodesNaf)
	reference.GET("/departements", getDepartements)
	reference.GET("/regions", getRegions)

	utils := router.Group("/utils", getAdminAuthMiddleware())
	utils.GET("/import", importHandler)
	utils.GET("/keycloak", getKeycloakUsers)
	utils.GET("/metrics", gin.WrapH(promhttp.Handler()))

	log.Print("Running API on " + viper.GetString("bind"))
	err := router.Run(viper.GetString("bind"))
	if err != nil {
		fmt.Println(err.Error())
	}
}

func getAdminAuthMiddleware() gin.HandlerFunc {
	var whitelist = viper.GetStringSlice("adminWhitelist")
	var wlmap = make(map[string]bool)
	for _, ip := range whitelist {
		wlmap[ip] = true
	}

	return func(c *gin.Context) {
		if !wlmap[c.ClientIP()] {
			log.Printf("Connection from %s is not granted in adminWhitelist, see config.toml\n", c.ClientIP())
			c.AbortWithStatus(403)
			return
		}
	}
}
