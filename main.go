package main

import (
	"fmt"
	"log"

	gocloak "github.com/Nerzal/gocloak/v6"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	pgx "github.com/jackc/pgx/v4"

	"github.com/spf13/viper"
)

var keycloak gocloak.GoCloak
var db *pgx.Conn
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

	// router.Use(dbMiddleware(db))

	if viper.GetBool("enableKeycloak") {
		router.Use(keycloakMiddleware)
	} else {
		router.Use(fakeCloakMiddleware)
	}

	router.GET("/entreprise/get/:siren", validSiren, getEntreprise)
	router.GET("/etablissement/get/:siret", validSiret, getEtablissement)
	router.GET("/etablissements/get/:siren", validSiren, getEntrepriseEtablissements)

	router.GET("/follow", getEtablissementsFollowedByCurrentUser)
	router.POST("/follow/:siret", validSiret, followEtablissement)
	router.DELETE("/follow/:siret", validSiret, unfollowEtablissement)

	router.GET("/etablissement/comments/:siret", validSiret, getEntrepriseComments)
	router.POST("/etablissement/comments/:siret", validSiret, addEntrepriseComment)

	router.GET("/listes", getListes)
	router.POST("/scores", getLastListeScores)
	router.POST("/scores/:id", getListeScores)

	router.GET("/reference/naf", getCodesNaf)
	router.GET("/reference/departements", getDepartements)
	router.GET("/reference/regions", getRegions)

	router.GET("/import", importHandler)

	log.Print("Running API on " + viper.GetString("bind"))
	err := router.Run(viper.GetString("bind"))
	if err != nil {
		fmt.Println(err.Error())
	}
}

// func dbMiddleware(db *pgx.Conn) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		c.Set("DB", db)
// 		c.Next()
// 	}
// }
