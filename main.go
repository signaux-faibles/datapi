package main

import (
	"database/sql"
	"fmt"
	"log"
	gocloak "github.com/Nerzal/gocloak/v5"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/spf13/viper"
)

var keycloak gocloak.GoCloak

func main() {
	loadConfig()
	db := connectDB()
	keycloak := connectKC()
	runAPI(db, keycloak)
}

func loadConfig() {
	viper.SetDefault("MigrationsDir", "./migrations")
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.ReadInConfig()
}

func runAPI(db *sql.DB, keycloak *gocloak.GoCloak) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AddAllowHeaders("Authorization")
	config.AddAllowMethods("GET", "POST")
	router.Use(cors.New(config))

	router.Use(dbMiddleware(db))
	//router.Use(keycloakMiddleware)

	router.GET("/entreprise/get/:siren", getEntreprise)
	router.GET("/etablissement/get/:siret", getEtablissement)
	router.GET("/etablissements/get/:siren", getEntrepriseEtablissements)

	router.GET("/follow", getEntreprisesFollowedByUser)
	router.POST("/follow/:siren", followEntreprise)

	router.GET("/entreprise/comments/:siren", getEntrepriseComments)
	router.POST("/entreprise/comments/:siren", addEntrepriseComment)

	router.GET("/listes", getListes)
	router.GET("/scores", getLastListeScores)
	router.GET("/scores/:liste", getListeScores)

	router.GET("/reference/naf", getCodesNaf)
	router.GET("/reference/departements", getDepartements)
	router.GET("/reference/regions", getRegions)

	log.Print("Running API on " + viper.GetString("bind"))
	err := router.Run(viper.GetString("bind"))
	if err != nil {
		fmt.Println(err.Error())
	}
}

func dummy(c *gin.Context) {
	log.Println("dummy")
	c.JSON(200, "dummy")
}

func dbMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("DB", db)
		c.Next()
	}
}