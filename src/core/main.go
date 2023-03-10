package core

import (
	"bytes"
	"context"
	"fmt"
	gocloak "github.com/Nerzal/gocloak/v10"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/datapi/src/db"
	"github.com/signaux-faibles/datapi/src/refresh"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/spf13/viper"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var keycloak gocloak.GoCloak
var mgoDB *mongo.Database
var wekanConfig WekanConfig

// StartDatapi se connecte aux bases de données et keycloak
func StartDatapi() {
	db.Init() // fail fast - on n'attend pas la première requête pour savoir si on peut se connecter à la db
	mgoDB = connectWekanDB()
	go watchWekanConfig(&wekanConfig, time.Second)
	keycloak = connectKC()
}

// LoadConfig charge la config toml
func LoadConfig(confDirectory, confFile, migrationDir string) {
	viper.SetDefault("MigrationsDir", migrationDir)
	viper.SetConfigName(confFile)
	viper.SetConfigType("toml")
	viper.AddConfigPath(confDirectory)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

// StartAPI expose l'api
func StartAPI() {
	if viper.GetBool("prod") {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = viper.GetStringSlice("corsAllowOrigins")
	config.AddAllowHeaders("Authorization")
	config.AddAllowMethods("GET", "POST", "DELETE")
	router.Use(cors.New(config))
	router.SetTrustedProxies(nil)

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
	follow.POST("", getCardsForCurrentUser)
	follow.POST("/:siret", validSiret, followEtablissement)
	follow.DELETE("/:siret", validSiret, unfollowEtablissement)

	export := router.Group("/export/", getKeycloakMiddleware(), logMiddleware)
	export.GET("/xlsx/follow", getXLSXFollowedByCurrentUser)
	export.POST("/xlsx/follow", getXLSXFollowedByCurrentUser)
	export.GET("/docx/follow", getDOCXFollowedByCurrentUser)
	export.POST("/docx/follow", getDOCXFollowedByCurrentUser)
	export.GET("/docx/siret/:siret", validSiret, getDOCXFromSiret)

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
	// utils.GET("/wekanImport", wekanImportHandler)
	utils.GET("/sireneImport", sireneImportHandler)
	utils.GET("/listImport/:algo", listImportHandler)

	refreshRoute := router.Group("/refresh", getKeycloakMiddleware(), logMiddleware)
	refreshRoute.GET("/start", refresh.StartHandler)
	refreshRoute.GET("/status/:uuid", refresh.StatusHandler)
	refreshRoute.GET("/last", refresh.LastHandler)
	refreshRoute.GET("/list/:status", refresh.ListHandler)

	wekan := router.Group("/wekan", getKeycloakMiddleware(), logMiddleware)
	wekan.GET("/cards/:siret", validSiret, wekanGetCardsHandler)
	wekan.POST("/cards/:siret", validSiret, wekanNewCardHandler)
	wekan.GET("/unarchive/:cardID", wekanUnarchiveCardHandler)
	wekan.GET("/join/:cardId", wekanJoinCardHandler)
	wekan.GET("/config", wekanConfigHandler)

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
	if c.Request.Body == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "request has nil body"})
		return
	}
	path := c.Request.URL.Path
	method := c.Request.Method
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var token []string
	if viper.GetBool("enableKeycloak") {
		token, err = getRawToken(c)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}
	} else {
		token = []string{"", "fakeKeycloak"}
	}

	_, err = db.Get().Exec(
		context.Background(),
		`insert into logs (path, method, body, token) values ($1, $2, $3, $4);`,
		path,
		method,
		string(body),
		token[1],
	)
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
