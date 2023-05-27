package main

import (
	"context"
	"datapi/pkg/campaign"
	"datapi/pkg/core"
	"datapi/pkg/db"
	"datapi/pkg/logPersistence"
	"datapi/pkg/kanban"
	"datapi/pkg/ops/imports"
	"datapi/pkg/ops/misc"
	"datapi/pkg/ops/refresh"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"log"
)

func main() {
	core.LoadConfig(".", "config", "./migrations")
	if viper.GetBool("prod") {
		gin.SetMode(gin.ReleaseMode)
	}
	ctx := context.Background()
	kanban := initWekanService(ctx)
	saver := logPersistence.NewPostgresLogSaver(ctx, db.Get())
	datapi, err := core.StartDatapi(kanban, saver.SaveLogToDB)
	if err != nil {
		log.Println("erreur pendant le démarrage de Datapi : ", err)
	}
	initAndStartAPI(datapi)
}

func initWekanService(ctx context.Context) core.KanbanService {
	return kanban.InitService(ctx,
		viper.GetString("wekanMgoURL"),
		viper.GetString("wekanMgoDB"),
		viper.GetString("wekanAdminUsername"),
		viper.GetString("wekanSlugDomainRegexp"))
}

func initAndStartAPI(datapi *core.Datapi) {
	router := gin.Default()
	datapi.InitAPI(router)
	core.AddEndpoint(router, "/ops/utils", misc.ConfigureEndpoint, core.AdminAuthMiddleware)
	core.AddEndpoint(router, "/ops/imports", imports.ConfigureEndpoint, core.AdminAuthMiddleware)
	core.AddEndpoint(router, "/ops/refresh", refresh.ConfigureEndpoint, core.AdminAuthMiddleware)
	core.AddEndpoint(router, "/campaign", campaign.ConfigureEndpoint)
	core.StartAPI(router)
}
