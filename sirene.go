package main

import (
	"context"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/goSirene"
	"github.com/spf13/viper"
)

func sireneImportHandler(c *gin.Context) {
	wg := sync.WaitGroup{}
	wg.Add(2)
	ctx, cancelCtx := context.WithCancel(context.Background())

	sireneUL := goSirene.SireneULParser(ctx, viper.GetString("sireneULpath"))
	go insertSireneUL(ctx, cancelCtx, wg, sireneUL)

	geoSireneFile, err := os.Open(viper.GetString("geoSirenePath"))
	if err == nil {
		geoSirene := goSirene.GeoSireneParser(ctx, geoSireneFile)
		go insertGeoSirene(ctx, cancelCtx, wg, geoSirene)
	}

	wg.Wait()

}

func insertSireneUL(ctx context.Context, cancelCtx context.CancelFunc, wg sync.WaitGroup, SireneUL chan goSirene.SireneUL) {
	cancelCtx()
	wg.Done()
}

func insertGeoSirene(ctx context.Context, cancelCtx context.CancelFunc, wg sync.WaitGroup, geoSirene chan goSirene.GeoSirene) {
	wg.Done()
}
