package main

import (
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/datapi/src/core"
	"github.com/signaux-faibles/datapi/src/ops/imports"
	"github.com/signaux-faibles/datapi/src/ops/misc"
	"github.com/signaux-faibles/datapi/src/ops/refresh"
	"github.com/spf13/viper"
)

func main() {
	core.LoadConfig(".", "config", "./migrations")
	if viper.GetBool("prod") {
		gin.SetMode(gin.ReleaseMode)
	}
	core.StartDatapi()
	initAndStartAPI()
}

func initAndStartAPI() {
	router := gin.Default()
	core.InitAPI(router)
	core.AddEndpoint(router, "ops/utils", misc.ConfigureEndpoint)
	core.AddEndpoint(router, "ops/imports", imports.ConfigureEndpoint)
	core.AddEndpoint(router, "ops/refresh", refresh.ConfigureEndpoint)
	core.StartAPI(router)
}
