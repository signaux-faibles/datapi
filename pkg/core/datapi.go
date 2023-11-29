// Package core embarque des types et fonctions communes de datapi
package core

import (
	"fmt"
	"log"
	"log/slog"
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
	saveAccessLog AccessLogSaver
	KanbanService KanbanService
}

// PrepareDatapi se connecte aux bases de données et keycloak
func PrepareDatapi(kanbanService KanbanService, saver AccessLogSaver) (*Datapi, error) {
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
		saveAccessLog: saver,
		KanbanService: kanbanService,
	}
	return &datapi, nil
}

// InitAPI initialise l'api
func (datapi *Datapi) InitAPI(router *gin.Engine) {
	config := cors.DefaultConfig()
	config.AllowOrigins = viper.GetStringSlice("corsAllowOrigins")
	config.AddExposeHeaders("Content-Disposition", "responseType", "Content-Type", "Cache-Control", "Connection", "Transfer-Encoding", "X-Accel-Buffering")

	config.AddAllowHeaders("Authorization", "responseType")
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
		slog.Warn("Attention : tentative de connexion à l'administration non autorisée", slog.String("ip", c.ClientIP()))
		c.AbortWithStatus(http.StatusForbidden)
	}
}

func configureKanbanEndpoint(path string, api *gin.Engine, handlers ...gin.HandlerFunc) {
	kanban := api.Group(path, handlers...)
	kanban.GET("/config", CheckAnyRolesMiddleware("wekan"), kanbanConfigHandler)
	kanban.GET("/cards/:siret", CheckAnyRolesMiddleware("wekan"), kanbanGetCardsHandler)
	kanban.POST("/follow", kanbanGetCardsForCurrentUserHandler)
	kanban.GET("/unarchive/:cardID", CheckAnyRolesMiddleware("wekan"), kanbanUnarchiveCardHandler)
	kanban.POST("/updateCard", CheckAnyRolesMiddleware("wekan"), kanbanUpdateCardHandler)
	kanban.POST("/card", CheckAnyRolesMiddleware("wekan"), kanbanNewCardHandler)
	kanban.GET("/card/join/:cardID", CheckAnyRolesMiddleware("wekan"), kanbanJoinCardHandler)
	kanban.GET("/card/part/:cardID", CheckAnyRolesMiddleware("wekan"), kanbanPartCardHandler)
	kanban.GET("/card/get/:cardID", CheckAnyRolesMiddleware("wekan"), kanbanGetCardHandler)
}

// True made global to ease pointers
var True = true

// False made global to ease pointers
var False = false

// EmptyString made global to ease pointers
var EmptyString = ""
