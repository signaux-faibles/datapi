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

	router.Use(keycloakMiddleware)

	router.GET("/entreprise/get/:siren", dummy)
	router.GET("/etablissement/get/:siret", dummy)
	router.GET("/etablissements/get/:siren", dummy)

	router.GET("/follow", dummy)
	router.POST("/follow/:siren", dummy)

	router.GET("/entreprise/comments/:siren", dummy)
	router.POST("/entreprise/comments/:siren", dummy)

	router.GET("/listes", dummy)
	router.POST("/scores", dummy)
	router.POST("/scores/:liste", dummy)

	router.POST("/reference", dummy)

	log.Print("Running API on " + viper.GetString("bind"))
	err := router.Run(viper.GetString("bind"))
	if err != nil {
		fmt.Println(err.Error())
	}
}

func dummy(c *gin.Context) {

}
