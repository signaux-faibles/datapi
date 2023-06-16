package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"datapi/pkg/core"
	"datapi/pkg/ops/imports"
	"datapi/pkg/ops/misc"
	"datapi/pkg/ops/refresh"
	"datapi/pkg/stats"
	"datapi/pkg/wekan"
)

func main() {
	core.LoadConfig(".", "config", "./migrations")
	if viper.GetBool("prod") {
		gin.SetMode(gin.ReleaseMode)
	}
	ctx := context.Background()
	kanban := initWekanService(ctx)
	saver, err := stats.NewPostgresLogSaverFromURL(context.Background(), viper.GetString("logs.db_url"))
	if err != nil {
		log.Fatal("erreur pendant l'instanciation du AccessLogSaver : ", err)
	}
	//_, err = stats.SyncPostgresLogSaver(saver, db.Get())
	//if err != nil {
	//	log.Printf("ERREUR : erreur pendant la synchronisation des logs : %s", err)
	//}
	datapi, err := core.StartDatapi(kanban, saver.SaveLogToDB)
	if err != nil {
		log.Println("erreur pendant le démarrage de Datapi : ", err)
	}
	if err != nil {
		log.Fatal("erreur pendant l'instanciation du LogSyncer : ", err)
	}
	initAndStartAPI(datapi)
}

func initWekanService(ctx context.Context) core.KanbanService {
	return wekan.InitService(ctx,
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
	core.StartAPI(router)
}
