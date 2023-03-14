package main

import (
	"github.com/signaux-faibles/datapi/src/core"
	"github.com/signaux-faibles/datapi/src/ops"
	"github.com/signaux-faibles/datapi/src/refresh"
)

func main() {
	core.LoadConfig(".", "config", "./migrations")
	core.StartDatapi()
	initAndStartAPI()
}

func initAndStartAPI() {
	api := core.InitAPI()
	core.AddEndpoint(api, "/refresh", refresh.ConfigureEndpoint)
	core.AddEndpoint(api, "/utils", ops.ConfigureEndpoint)
	core.StartAPI(api)
}
