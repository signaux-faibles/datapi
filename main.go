package main

import (
	"github.com/signaux-faibles/datapi/src/core"
)

func main() {
	core.LoadConfig(".", "config", "./migrations")
	core.StartDatapi()
	core.RunAPI()
}
