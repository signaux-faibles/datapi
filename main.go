package main

import (
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/datapi/src/core"
	"github.com/signaux-faibles/datapi/src/ops"
	"github.com/signaux-faibles/datapi/src/refresh"
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
	core.AddEndpoint(router, "/refresh", refresh.ConfigureEndpoint)
	core.AddEndpoint(router, "/utils", ops.ConfigureEndpoint)
	core.StartAPI(router)
}
