// Package core embarque des types et fonctions communes de datapi
package core

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"time"

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
	router.Use(gin.Recovery())
	config := cors.DefaultConfig()
	config.AllowOrigins = viper.GetStringSlice("corsAllowOrigins")
	config.AddExposeHeaders("Content-Disposition", "responseType", "Content-Type", "Cache-Control", "Connection", "Transfer-Encoding", "X-Accel-Buffering")

	config.AddAllowHeaders("Authorization", "responseType")
	config.AddAllowMethods("GET", "POST", "DELETE")
	router.Use(cors.New(config))
	const (
		IP_REVERSE_PROXY       string = "10.0.2.100"
		IP_REVERSE_PROXY_ADMIN string = "10.3.0.195"
	)
	err := router.SetTrustedProxies([]string{IP_REVERSE_PROXY, IP_REVERSE_PROXY_ADMIN})

	router.Use(ExtractEmailFromJWT())
	router.Use(ExtractBodyFromHttpJson())
	router.Use(CustomLogger())

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
	etablissement.POST("/search/total", searchEtablissementTotalHandler)

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

	configureKanbanEndpoint("/kanban", router, AuthMiddleware(), datapi.LogMiddleware)
}

func CustomLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		skipPaths := map[string]bool{
			"/healthcheck":       true,
			"/ops/utils/metrics": true,
		}
		if skipPaths[param.Path] {
			return ""
		}
		email := param.Keys["email"]
		body := param.Keys["body"]
		return fmt.Sprintf("[GIN]\t%s\t|%d\t|%s\t|%s\t|%v\t|%s %s%s\n",
			param.TimeStamp.Format(time.DateTime),
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			email,
			param.Method,
			param.Path,
			body,
		)
	})
}

func ExtractEmailFromJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			tokenStr := strings.TrimPrefix(auth, "Bearer ")
			parser := jwt.NewParser()
			token, _, err := parser.ParseUnverified(tokenStr, jwt.MapClaims{})
			if err == nil {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if email, exists := claims["email"]; exists {
						c.Set("email", email)
					}
				}
			}
		}
		c.Next()
	}
}

func ExtractBodyFromHttpJson() gin.HandlerFunc {
	return func(c *gin.Context) {
		body := ""
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // recharge le body inchangé

			// transforme ["A", "B", "C"] en "[A, B, C]" pour la lisibilité des logs
			var obj map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &obj); err == nil {
				for k, v := range obj {
					if arr, ok := v.([]interface{}); ok {
						strs := make([]string, len(arr))
						for i, val := range arr {
							strs[i] = fmt.Sprintf("%v", val)
						}
						obj[k] = "[" + strings.Join(strs, ", ") + "]"
					}
				}
				pretty, _ := json.MarshalIndent(obj, "", "  ")
				body = string(pretty)
			}
		}
		c.Set("body", body)
		c.Next()
	}
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

	listOnly := flag.Bool("list-routes", false, "List all registered routes and exit")
	flag.Parse()

	if *listOnly {
		utils.ListRoutes(router)
	} else {
		err := router.Run(addr)

		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

// AuthMiddleware définit le middleware qui gère l'authentification
func AuthMiddleware() gin.HandlerFunc {
	if viper.GetBool("enableKeycloak") {
		return keycloakMiddleware
	}
	return fakeCloakMiddleware
}

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
	kanban.GET("/card/membersHistory/:cardID", CheckAnyRolesMiddleware("wekan"), kanbanGetCardMembersHistoryHandler)
}

// True made global to ease pointers
var True = true

// False made global to ease pointers
var False = false

// EmptyString made global to ease pointers
var EmptyString = ""
