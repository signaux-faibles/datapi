package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/goSirene"
	"github.com/spf13/viper"
)

func sireneImportHandler(c *gin.Context) {
	if viper.GetString("sireneULPath") == "" || viper.GetString("geoSirenePath") == "" {
		c.AbortWithStatusJSON(409, "not supported, missing parameters in server configuration")
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	ctx, cancelCtx := context.WithCancel(context.Background())
	//go upsertSireneUL(ctx, cancelCtx, &wg)
	go upsertGeoSirene(ctx, cancelCtx, &wg)
	time.Sleep(time.Second * 5)
	cancelCtx()
	wg.Wait()
}

func upsertGeoSirene(ctx context.Context, cancelCtx context.CancelFunc, wg *sync.WaitGroup) {
	file, err := os.Open(viper.GetString("sireneULPath"))
	if err != nil {
		cancelCtx()
		return
	}
	geoSirene := goSirene.GeoSireneParser(ctx, file)
	for i := range geoSirene {
		fmt.Println(i)
	}
	wg.Done()
}

func upsertSireneUL(ctx context.Context, cancelCtx context.CancelFunc, wg *sync.WaitGroup) {
	sireneUL := goSirene.SireneULParser(ctx, viper.GetString("sireneULPath"))
	count := 0
	for sirene := range sireneUL {
		count++
		fmt.Println(sirene)
		if count == 5000 {
			// fmt.Println(getInsertSireneUL(5))
			count = 0
		}
	}
	wg.Done()
}

func getInsertSireneUL(length int) [2]string {
	var create = `
		create temporary table tmp_sirene_ul 
		(siren text,
		siret_siege text, 
		raison_sociale text, 
		prenom1 text, 
		prenom2 text, 
		prenom3 text, 
		prenom4 text, 
		nom text,
		nom_usage text, 
		statut_juridique text, 
		creation date, 
		sigle text, 
		identification_association text, 
		tranche_effetif text,
		annee_effectif text, 
		categorie text, 
		annee_categorie text, 
		etat_administratif text, 
		nomusage text, 
		denomination text, 
		economie_sociale_solidaire boolean,
		caractere_employeur boolean)
		on commit drop;`
	sql := `insert into tmp_sirene_url (siren text,
		siret_siege text, 
		raison_sociale text, 
		prenom1 text, 
		prenom2 text, 
		prenom3 text, 
		prenom4 text, 
		nom text,
		nom_usage text, 
		statut_juridique text, 
		creation date, 
		sigle text, 
		identification_association text, 
		tranche_effetif text,
		annee_effectif text, 
		categorie text, 
		annee_categorie text, 
		etat_administratif text, 
		nomusage text, 
		denomination text, 
		economie_sociale_solidaire boolean,
		caractere_employeur boolean) values `
	values := []string{}
	for i := 0; i < length; i++ {
		anchors := ids(i, 22)
		values = append(values, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)", anchors...))
	}
	sql = sql + strings.Join(values, ",")
	return [2]string{create, sql}
}

func ids(page int, length int) []interface{} {
	r := []interface{}{}
	for i := page*length + 1; i <= (page+1)*length; i++ {
		r = append(r, i)
	}
	return r
}

func insertGeoSirene(ctx context.Context, cancelCtx context.CancelFunc, wg sync.WaitGroup, geoSirene chan goSirene.GeoSirene) {
	wg.Done()
}
