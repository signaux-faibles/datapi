package wekan

import (
	"context"
	"log"
	"time"
)

func watchWekanConfig(period time.Duration) {
	for {
		loadWekanConfig()
		time.Sleep(period)
	}
}

func loadWekanConfig() {
	newWekanConfig, err := wekan.SelectConfig(context.Background())
	if err != nil {
		log.Printf("Erreur lors du chargement de la config wekan pour l'utilisateur : %s", err)
	} else {
		wekanConfigMutex.Lock()
		wekanConfig = newWekanConfig
		wekanConfigMutex.Unlock()
	}
}
