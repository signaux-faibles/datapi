package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/datapi/src/core"
	"github.com/signaux-faibles/datapi/src/ops/imports"
	"github.com/signaux-faibles/datapi/src/ops/misc"
	"github.com/signaux-faibles/datapi/src/ops/refresh"
	"github.com/signaux-faibles/datapi/src/wekan"
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
	err := core.StartDatapi(kanban)
	if err != nil {
		log.Println("erreur pendant le démarrage de Datapi : ", err)
	}
	initAndStartAPI()
}

func initWekanService(ctx context.Context) core.KanbanService {
	return wekan.InitService(ctx,
		viper.GetString("wekanMgoURL"),
		viper.GetString("wekanMgoDB"),
		viper.GetString("wekanAdminUsername"),
		viper.GetString("wekanSlugDomainRegexp"))
}

func initAndStartAPI() {
	router := gin.Default()
	core.InitAPI(router)
	core.AddEndpoint(router, "/ops/utils", misc.ConfigureEndpoint)
	core.AddEndpoint(router, "/ops/imports", imports.ConfigureEndpoint)
	core.AddEndpoint(router, "/ops/refresh", refresh.ConfigureEndpoint)
	core.StartAPI(router)
}
