package main

import (
	"github.com/signaux-faibles/datapi/src/core"
	"github.com/signaux-faibles/datapi/src/refresh"
)

func main() {
	core.LoadConfig(".", "config", "./migrations")
	core.StartDatapi()
	api := core.InitAPI()
	core.ConfigureAPI(api, refresh.ConfigureEndpoint)
	core.StartAPI(api)
}
