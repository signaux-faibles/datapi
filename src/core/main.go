package core

import (
	"bytes"
	"context"
	"fmt"
	gocloak "github.com/Nerzal/gocloak/v10"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/datapi/src/db"
	"github.com/signaux-faibles/datapi/src/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

var keycloak gocloak.GoCloak
var mgoDB *mongo.Database
var wekanConfig WekanConfig

// Endpoint handler pour définir un endpoint sur gin
type Endpoint func(path string, api *gin.Engine)

// StartDatapi se connecte aux bases de données et keycloak
func StartDatapi() {
	db.Init() // fail fast - on n'attend pas la première requête pour savoir si on peut se connecter à la db
	mgoDB = connectWekanDB()
	go watchWekanConfig(&wekanConfig, time.Second)
	keycloak = connectKC()
}

// LoadConfig charge la config toml
func LoadConfig(confDirectory, confFile, migrationDir string) {
	viper.SetDefault("migrationsDir", migrationDir)
	viper.SetConfigName(confFile)
	viper.SetConfigType("toml")
	viper.AddConfigPath(confDirectory)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

// InitAPI initialise l'api
func InitAPI() *gin.Engine {
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

	entreprise := router.Group("/entreprise", AuthMiddleware(), LogMiddleware)
	entreprise.GET("/viewers/:siren", validSiren, getEntrepriseViewers)
	entreprise.GET("/get/:siren", validSiren, getEntreprise)
	entreprise.GET("/all/:siren", validSiren, getEntrepriseEtablissements)

	etablissement := router.Group("/etablissement", AuthMiddleware(), LogMiddleware)
	etablissement.GET("/viewers/:siret", validSiret, getEtablissementViewers)
	etablissement.GET("/get/:siret", validSiret, getEtablissement)
	etablissement.GET("/comments/:siret", validSiret, getEntrepriseComments)
	etablissement.POST("/comments/:siret", validSiret, addEntrepriseComment)
	etablissement.PUT("/comments/:id", updateEntrepriseComment)
	etablissement.POST("/search", searchEtablissementHandler)

	follow := router.Group("/follow", AuthMiddleware(), LogMiddleware)
	follow.GET("", getEtablissementsFollowedByCurrentUser)
	follow.POST("", getCardsForCurrentUser)
	follow.POST("/:siret", validSiret, followEtablissement)
	follow.DELETE("/:siret", validSiret, unfollowEtablissement)

	export := router.Group("/export/", AuthMiddleware(), LogMiddleware)
	export.GET("/xlsx/follow", getXLSXFollowedByCurrentUser)
	export.POST("/xlsx/follow", getXLSXFollowedByCurrentUser)
	export.GET("/docx/follow", getDOCXFollowedByCurrentUser)
	export.POST("/docx/follow", getDOCXFollowedByCurrentUser)
	export.GET("/docx/siret/:siret", validSiret, getDOCXFromSiret)

	listes := router.Group("/listes", AuthMiddleware(), LogMiddleware)
	listes.GET("", getListes)

	scores := router.Group("/scores", AuthMiddleware(), LogMiddleware)
	scores.POST("/liste", getLastListeScores)
	scores.POST("/liste/:id", getListeScores)
	scores.POST("/xls/:id", getXLSListeScores)

	reference := router.Group("/reference", AuthMiddleware(), LogMiddleware)
	reference.GET("/naf", getCodesNaf)
	reference.GET("/departements", getDepartements)
	reference.GET("/regions", getRegions)

	fce := router.Group("/fce", AuthMiddleware(), LogMiddleware)
	fce.GET("/:siret", validSiret, getFceURL)

	wekan := router.Group("/wekan", AuthMiddleware(), LogMiddleware)
	wekan.GET("/cards/:siret", validSiret, wekanGetCardsHandler)
	wekan.POST("/cards/:siret", validSiret, wekanNewCardHandler)
	wekan.GET("/unarchive/:cardID", wekanUnarchiveCardHandler)
	wekan.GET("/join/:cardId", wekanJoinCardHandler)
	wekan.GET("/config", wekanConfigHandler)
	return router
}

// AddEndpoint permet de rajouter un endpoint au niveau de l'API
func AddEndpoint(router *gin.Engine, path string, endpoint Endpoint) {
	endpoint(path, router)
}

// StartAPI : démarre le serveur
func StartAPI(router *gin.Engine) {
	log.Print("Running API on " + viper.GetString("bind"))
	err := router.Run(viper.GetString("bind"))
	if err != nil {
		fmt.Println(err.Error())
	}
}

// AuthMiddleware définit le middleware qui gère l'authentification
func AuthMiddleware() gin.HandlerFunc {
	if viper.GetBool("enableKeycloak") {
		return keycloakMiddleware
	}
	return fakeCloakMiddleware
}

// LogMiddleware définit le middleware qui gère les logs
func LogMiddleware(c *gin.Context) {
	if c.Request.Body == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "request has nil body"})
		return
	}
	path := c.Request.URL.Path
	method := c.Request.Method
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var token []string
	if viper.GetBool("enableKeycloak") {
		token, err = getRawToken(c)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
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
