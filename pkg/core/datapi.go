// Package core embarque des types et fonctions communes de datapi
package core

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Nerzal/gocloak/v10"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"datapi/pkg/db"
	"datapi/pkg/utils"
)

var keycloak gocloak.GoCloak

var Regions map[Region][]CodeDepartement
var Departements map[CodeDepartement]string

// Endpoint handler pour définir un endpoint sur gin
type Endpoint func(endpoint *gin.RouterGroup)

var Kanban KanbanService

type Datapi struct {
	saveAPICall SaveLogInfos
}

// StartDatapi se connecte aux bases de données et keycloak
func StartDatapi(kanbanService KanbanService, saver SaveLogInfos) (*Datapi, error) {
	var err error
	db.Init() // fail fast - on n'attend pas la première requête pour savoir si on peut se connecter à la db
	Departements, err = loadDepartementReferentiel()
	if err != nil {
		return nil, fmt.Errorf("erreur pendant le chargement du référentiel des départements : %w", err)
	}
	Regions, err = loadRegionsReferentiel()
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("erreur pendant le chargement du référentiel des régions : %w", err)
	}

	if kanbanService == nil {
		return nil, fmt.Errorf("le service Kanban n'est pas paramétré")
	}
	Kanban = kanbanService
	keycloak = connectKC()

	datapi := Datapi{
		saveAPICall: saver,
	}
	return &datapi, nil
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
func (datapi *Datapi) InitAPI(router *gin.Engine) {
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

	entreprise := router.Group("/entreprise", AuthMiddleware(), datapi.LogMiddleware)
	entreprise.GET("/viewers/:siren", checkSirenFormat, getEntrepriseViewers)
	entreprise.GET("/get/:siren", checkSirenFormat, getEntreprise)
	entreprise.GET("/all/:siren", checkSirenFormat, getEntrepriseEtablissements)

	etablissement := router.Group("/etablissement", AuthMiddleware(), datapi.LogMiddleware)
	etablissement.GET("/viewers/:siret", checkSiretFormat, getEtablissementViewers)
	etablissement.GET("/get/:siret", checkSiretFormat, getEtablissement)
	etablissement.GET("/comments/:siret", checkSiretFormat, getEntrepriseComments)
	etablissement.POST("/comments/:siret", checkSiretFormat, addEntrepriseComment)
	etablissement.PUT("/comments/:id", updateEntrepriseComment)
	etablissement.POST("/search", searchEtablissementHandler)

	follow := router.Group("/follow", AuthMiddleware(), datapi.LogMiddleware)
	follow.GET("", getEtablissementsFollowedByCurrentUser)
	follow.POST("/:siret", checkSiretFormat, followEtablissement)
	follow.DELETE("/:siret", checkSiretFormat, unfollowEtablissement)

	export := router.Group("/export/", AuthMiddleware(), datapi.LogMiddleware)
	export.POST("/xlsx/follow", getXLSXFollowedByCurrentUser)
	export.POST("/docx/follow", getDOCXFollowedByCurrentUser)
	export.GET("/docx/siret/:siret", checkSiretFormat, getDOCXFromSiret)

	listes := router.Group("/listes", AuthMiddleware(), datapi.LogMiddleware)
	listes.GET("", getListes)

	scores := router.Group("/scores", AuthMiddleware(), datapi.LogMiddleware)
	scores.POST("/liste", getLastListeScores)
	scores.POST("/liste/:id", getListeScores)
	scores.POST("/xls/:id", getXLSListeScores)

	reference := router.Group("/reference", AuthMiddleware(), datapi.LogMiddleware)
	reference.GET("/naf", getCodesNaf)
	reference.GET("/departements", departementsHandler)
	reference.GET("/regions", regionsHandler)

	fce := router.Group("/fce", AuthMiddleware(), datapi.LogMiddleware)
	fce.GET("/:siret", checkSiretFormat, getFceURL)

	configureKanbanEndpoint("/kanban", router, AuthMiddleware(), datapi.LogMiddleware)
}

// AddEndpoint permet de rajouter un endpoint au niveau de l'API
func AddEndpoint(router *gin.Engine, path string, endpoint Endpoint, handlers ...gin.HandlerFunc) {
	group := router.Group(path, handlers...)
	endpoint(group)
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

const forbiddenIPMessage = "Erreur : Une tentative de connexion depuis `%s`, ce qui n'est pas autorisé dans `adminWhitelist`, voir config.toml\n"

// AdminAuthMiddleware stoppe la requête si l'ip client n'est pas contenue dans la whitelist
func AdminAuthMiddleware(c *gin.Context) {
	ok := utils.AcceptIP(c.ClientIP())
	if !ok {
		log.Printf(forbiddenIPMessage, c.ClientIP())
		c.AbortWithStatus(http.StatusForbidden)
	}
}

func configureKanbanEndpoint(path string, api *gin.Engine, handlers ...gin.HandlerFunc) {
	kanban := api.Group(path, handlers...)
	kanban.GET("/config", CheckAnyRolesMiddleware("wekan"), kanbanConfigHandler)
	kanban.GET("/cards/:siret", CheckAnyRolesMiddleware("wekan"), kanbanGetCardsHandler)
	kanban.POST("/follow", kanbanGetCardsForCurrentUserHandler)
	kanban.POST("/card", CheckAnyRolesMiddleware("wekan"), kanbanNewCardHandler)
	kanban.GET("/unarchive/:cardID", CheckAnyRolesMiddleware("wekan"), kanbanUnarchiveCardHandler)
}

// True made global to ease pointers
var True = true

// False made global to ease pointers
var False = false

// EmptyString made global to ease pointers
var EmptyString = ""
