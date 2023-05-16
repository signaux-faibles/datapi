// Package core embarque des types et fonctions communes de datapi
package core

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Nerzal/gocloak/v10"
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

var oldWekanConfig OldWekanConfig

var Regions map[Region][]CodeDepartement
var Departements map[CodeDepartement]string

// Endpoint handler pour définir un endpoint sur gin
type Endpoint func(path string, api *gin.Engine)

var kanban KanbanService

// StartDatapi se connecte aux bases de données et keycloak
func StartDatapi(kanbanService KanbanService) error {
	var err error
	db.Init() // fail fast - on n'attend pas la première requête pour savoir si on peut se connecter à la db
	Departements, err = loadDepartementReferentiel()
	if err != nil {
		return fmt.Errorf("erreur pendant le chargement du référentiel des départements : %w", err)
	}
	Regions, err = loadRegionsReferentiel()
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("erreur pendant le chargement du référentiel des régions : %w", err)
	}

	// TODO supprimer
	mgoDB = connectWekanDB()
	go watchOldWekanConfig(&oldWekanConfig, time.Minute)
	if kanbanService == nil {
		return fmt.Errorf("le service Kanban n'est pas paramétré")
	}
	kanban = kanbanService
	keycloak = connectKC()
	return nil
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
func InitAPI(router *gin.Engine) {

	config := cors.DefaultConfig()
	config.AllowOrigins = viper.GetStringSlice("corsAllowOrigins")
	config.AddExposeHeaders("Content-Disposition")

	config.AddAllowHeaders("Authorization")
	config.AddAllowMethods("GET", "POST", "DELETE")
	router.Use(cors.New(config))
	err := router.SetTrustedProxies(nil)
	if err != nil {
		panic(err.Error())
	}

	entreprise := router.Group("/entreprise", AuthMiddleware(), LogMiddleware)
	entreprise.GET("/viewers/:siren", checkSirenFormat, getEntrepriseViewers)
	entreprise.GET("/get/:siren", checkSirenFormat, getEntreprise)
	entreprise.GET("/all/:siren", checkSirenFormat, getEntrepriseEtablissements)

	etablissement := router.Group("/etablissement", AuthMiddleware(), LogMiddleware)
	etablissement.GET("/viewers/:siret", checkSiretFormat, getEtablissementViewers)
	etablissement.GET("/get/:siret", checkSiretFormat, getEtablissement)
	etablissement.GET("/comments/:siret", checkSiretFormat, getEntrepriseComments)
	etablissement.POST("/comments/:siret", checkSiretFormat, addEntrepriseComment)
	etablissement.PUT("/comments/:id", updateEntrepriseComment)
	etablissement.POST("/search", searchEtablissementHandler)

	follow := router.Group("/follow", AuthMiddleware(), LogMiddleware)
	follow.GET("", getEtablissementsFollowedByCurrentUser)
	follow.POST("/:siret", checkSiretFormat, followEtablissement)
	follow.DELETE("/:siret", checkSiretFormat, unfollowEtablissement)

	export := router.Group("/export/", AuthMiddleware(), LogMiddleware)
	export.GET("/xlsx/follow", getXLSXFollowedByCurrentUser)
	export.POST("/xlsx/follow", getXLSXFollowedByCurrentUser)
	export.GET("/docx/follow", getDOCXFollowedByCurrentUser)
	export.POST("/docx/follow", getDOCXFollowedByCurrentUser)
	export.GET("/docx/siret/:siret", checkSiretFormat, getDOCXFromSiret)

	listes := router.Group("/listes", AuthMiddleware(), LogMiddleware)
	listes.GET("", getListes)

	scores := router.Group("/scores", AuthMiddleware(), LogMiddleware)
	scores.POST("/liste", getLastListeScores)
	scores.POST("/liste/:id", getListeScores)
	scores.POST("/xls/:id", getXLSListeScores)

	reference := router.Group("/reference", AuthMiddleware(), LogMiddleware)
	reference.GET("/naf", getCodesNaf)
	reference.GET("/departements", departementsHandler)
	reference.GET("/regions", regionsHandler)

	fce := router.Group("/fce", AuthMiddleware(), LogMiddleware)
	fce.GET("/:siret", checkSiretFormat, getFceURL)

	configureKanbanEndpoint("/kanban", router)
}

// AddEndpoint permet de rajouter un endpoint au niveau de l'API
func AddEndpoint(router *gin.Engine, path string, endpoint Endpoint) {
	endpoint(path, router)
}

// StartAPI : démarre le serveur
func StartAPI(router *gin.Engine) {
	addr := viper.GetString("bind")
	log.Print("Running API on " + addr)
	err := router.Run(addr)
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

const forbiddenIPMessage = "Erreur : Une tentative de connexion depuis `%s`, ce qui n'est pas autorisé dans `adminWhitelist`, voir config.toml\n"

// AdminAuthMiddleware stoppe la requête si l'ip client n'est pas contenue dans la whitelist
func AdminAuthMiddleware(c *gin.Context) {
	ok := utils.AcceptIP(c.ClientIP())
	if !ok {
		log.Printf(forbiddenIPMessage, c.ClientIP())
		c.AbortWithStatus(http.StatusForbidden)
	}
}

func configureKanbanEndpoint(path string, api *gin.Engine) {
	kanban := api.Group(path, AuthMiddleware(), CheckAnyRolesMiddleware("wekan"), LogMiddleware)
	kanban.GET("/config", kanbanConfigHandler)
	kanban.GET("/cards/:siret", kanbanGetCardsHandler)
	kanban.POST("/follow", kanbanGetCardsForCurrentUserHandler)
	kanban.POST("/card", kanbanNewCardHandler)
	kanban.GET("/unarchive/:cardID", kanbanUnarchiveCardHandler)
}

// True made global to ease pointers
var True = true

// False made global to ease pointers
var False = false

// EmptyString made global to ease pointers
var EmptyString = ""
