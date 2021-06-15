package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
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
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func runAPI() {
	if viper.GetBool("prod") {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:8081"}
	config.AddAllowHeaders("Authorization")
	config.AddAllowMethods("GET", "POST", "DELETE")
	router.Use(cors.New(config))

	entreprise := router.Group("/entreprise", getKeycloakMiddleware(), logMiddleware)
	entreprise.GET("/viewers/:siren", validSiren, getEntrepriseViewers)
	entreprise.GET("/get/:siren", validSiren, getEntreprise)
	entreprise.GET("/all/:siren", validSiren, getEntrepriseEtablissements)

	etablissement := router.Group("/etablissement", getKeycloakMiddleware(), logMiddleware)
	etablissement.GET("/viewers/:siret", validSiret, getEtablissementViewers)
	etablissement.GET("/get/:siret", validSiret, getEtablissement)
	etablissement.GET("/comments/:siret", validSiret, getEntrepriseComments)
	etablissement.POST("/comments/:siret", validSiret, addEntrepriseComment)
	etablissement.PUT("/comments/:id", updateEntrepriseComment)
	etablissement.POST("/search", searchEtablissementHandler)

	follow := router.Group("/follow", getKeycloakMiddleware(), logMiddleware)
	follow.GET("", getEtablissementsFollowedByCurrentUser)
	follow.POST("/:siret", validSiret, followEtablissement)
	follow.DELETE("/:siret", validSiret, unfollowEtablissement)
	follow.GET("/xls", getXLSFollowedByCurrentUser)
	follow.GET("/docx", getDOCXFollowedByCurrentUser)

	listes := router.Group("/listes", getKeycloakMiddleware(), logMiddleware)
	listes.GET("", getListes)

	scores := router.Group("/scores", getKeycloakMiddleware(), logMiddleware)
	scores.POST("/liste", getLastListeScores)
	scores.POST("/liste/:id", getListeScores)
	scores.POST("/xls/:id", getXLSListeScores)

	reference := router.Group("/reference", getKeycloakMiddleware(), logMiddleware)
	reference.GET("/naf", getCodesNaf)
	reference.GET("/departements", getDepartements)
	reference.GET("/regions", getRegions)

	fce := router.Group("/fce", getKeycloakMiddleware(), logMiddleware)
	fce.GET("/:siret", validSiret, getFceURL)

	utils := router.Group("/utils", getAdminAuthMiddleware())
	utils.GET("/import", importHandler)
	utils.GET("/keycloak", getKeycloakUsers)
	utils.GET("/metrics", gin.WrapH(promhttp.Handler()))
	utils.GET("/wekanImport", wekanImportHandler)
	utils.GET("/wekanListCards", wekanGetListCardsHandler)

	wekan := router.Group("/wekan", getKeycloakMiddleware(), logMiddleware)
	wekan.GET("/cards/:siret", wekanGetCardHandler)
	wekan.POST("/cards/:siret", wekanNewCardHandler)

	log.Print("Running API on " + viper.GetString("bind"))
	err := router.Run(viper.GetString("bind"))
	if err != nil {
		fmt.Println(err.Error())
	}
}

func getKeycloakMiddleware() gin.HandlerFunc {
	if viper.GetBool("enableKeycloak") {
		return keycloakMiddleware
	}
	return fakeCloakMiddleware
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

func logMiddleware(c *gin.Context) {
	path := c.Request.URL.Path
	method := c.Request.Method
	body, _ := ioutil.ReadAll(c.Request.Body)
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	var token []string
	var err error
	if viper.GetBool("enableKeycloak") {
		token, err = getRawToken(c)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}
	} else {
		token = []string{"", "fakeKeycloak"}
	}

	_, err = db.Exec(context.Background(), `insert into logs (path, method, body, token) 
	values ($1, $2, $3, $4);`, path, method, string(body), token[1])
	if err != nil {
		c.AbortWithStatus(500)
		return
	}

	c.Next()
}

// True made global to ease pointers
var True = true

// False made global to ease pointers
var False = false

// EmptyString made global to ease pointers
var EmptyString = ""
