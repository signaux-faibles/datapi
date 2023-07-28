package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"datapi/pkg/campaign"
	"datapi/pkg/core"
	"datapi/pkg/kanban"
	"datapi/pkg/ops/imports"
	"datapi/pkg/ops/misc"
	"datapi/pkg/ops/refresh"
	"datapi/pkg/stats"
)

func main() {
	core.LoadConfig(".", "config", "./migrations")
	if viper.GetBool("prod") {
		gin.SetMode(gin.ReleaseMode)
	}
	ctx := context.Background()
	kanbanService := initWekanService(ctx)
	saver := buildLogSaver()
	statsAPI := buildStatsAPI()

	datapi, err := core.PrepareDatapi(kanbanService, saver.SaveLogToDB)
	if err != nil {
		log.Println("erreur pendant le démarrage de Datapi : ", err)
	}
	initAndStartAPI(datapi, statsAPI)
}

func initWekanService(ctx context.Context) core.KanbanService {
	return kanban.InitService(ctx,
		viper.GetString("wekanMgoURL"),
		viper.GetString("wekanMgoDB"),
		viper.GetString("wekanAdminUsername"),
		viper.GetString("wekanSlugDomainRegexp"))
}

func initAndStartAPI(datapi *core.Datapi, statsAPI *stats.API) {
	router := gin.Default()
	datapi.InitAPI(router)
	core.AddEndpoint(router, "/ops/utils", misc.ConfigureEndpoint, core.AdminAuthMiddleware)
	core.AddEndpoint(router, "/ops/imports", imports.ConfigureEndpoint, core.AdminAuthMiddleware)
	core.AddEndpoint(router, "/ops/refresh", refresh.ConfigureEndpoint, core.AdminAuthMiddleware)
	core.AddEndpoint(router, "/campaign", campaign.ConfigureEndpoint(datapi.KanbanService), core.AuthMiddleware(), datapi.LogMiddleware)
	core.AddEndpoint(router, "/stats", statsAPI.ConfigureEndpoint, core.AuthMiddleware(), datapi.LogMiddleware)
	core.StartAPI(router)
}

func buildLogSaver() *stats.PostgresLogSaver {
	saver, err := stats.NewPostgresLogSaverFromURL(context.Background(), viper.GetString("logs.db_url"))
	if err != nil {
		log.Fatal("erreur pendant l'instanciation du AccessLogSaver : ", err)
	}
	return saver
}

func buildStatsAPI() *stats.API {
	statsModule, err := stats.NewAPIFromConfiguration(context.Background(), viper.GetString("logs.db_url"))
	if err != nil {
		log.Fatal("erreur pendant l'instanciation du StatsAPI : ", err)
	}
	return statsModule
}
