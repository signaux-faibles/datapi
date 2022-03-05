package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/btree"
	"github.com/jackc/pgx/v4"
	"github.com/signaux-faibles/goSirene"
	"github.com/spf13/viper"
)

func sireneImportHandler(c *gin.Context) {
	if viper.GetString("sireneULPath") == "" || viper.GetString("geoSirenePath") == "" {
		c.AbortWithStatusJSON(409, "not supported, missing parameters in server configuration")
		return
	}
	log.Println("Truncate etablissement & entreprise table")
	tx, err := db.Begin(context.Background())
	defer tx.Commit(context.Background())
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	err = truncateSirens(tx)
	log.Println("Tables truncated")
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	ctx, cancelCtx := context.WithCancel(context.Background())
	go insertSireneUL(ctx, cancelCtx, &wg)
	go insertGeoSirene(ctx, cancelCtx, &wg)

	wg.Wait()
}

func insertGeoSirene(ctx context.Context, cancelCtx context.CancelFunc, wg *sync.WaitGroup) {
	file, err := os.Open(viper.GetString("geoSirenePath"))
	if err != nil {
		cancelCtx()
		return
	}
	geoSirene := goSirene.GeoSireneParser(ctx, file)
	count := 0
	var batch []goSirene.GeoSirene
	for sirene := range geoSirene {
		count++
		if count == 100000 {
			count = 0
			err := sqlGeoSirene(ctx, batch)
			if err != nil {
				cancelCtx()
				fmt.Println(err.Error())
				continue
			}
			batch = []goSirene.GeoSirene{}
			log.Println("upsert GeoSirene")
		}
		if sirene.Error() == nil {
			batch = append(batch, sirene)
		}
	}
	err = sqlGeoSirene(ctx, batch)
	if err != nil {
		cancelCtx()
		fmt.Println(err.Error())
	}
	wg.Done()
}

func sqlGeoSirene(ctx context.Context, data []goSirene.GeoSirene) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	tx.Exec(ctx, `
	create temporary table tmp_geosirene 
	(siret text, 
		siren text, 
		siege bool, 
		creation timestamp, 
		complement_adresse text, 
		numero_voie text,
		indice_repetition text, 
		type_voie text, 
		voie text, 
		commune text, 
		commune_etranger text, 
		distribution_speciale text,
		code_commune text, 
		code_cedex text, 
		cedex text, 
		code_pays_etranger text, 
		pays_etranger text, 
		code_postal text,
		departement text, 
		code_activite text, 
		nomen_activite text, 
		latitude real, 
		longitude real, 
		tranche_effectif text, 
		annee_effectif text, 
		code_activite_registre_metiers text,
		etat_administratif text, 
		enseigne text, 
		denomination_usuelle text, 
		caractere_employeur bool,
		insert bool)
	on commit drop;`)

	batch := &pgx.Batch{}
	sql := `
	  insert into tmp_geosirene (siret, siren, siege, creation, complement_adresse, numero_voie,
			indice_repetition, type_voie, voie, commune, commune_etranger, distribution_speciale,
			code_commune, code_cedex, cedex, code_pays_etranger, pays_etranger, code_postal,
			departement, code_activite, nomen_activite, latitude, longitude, 
		  tranche_effectif, annee_effectif, code_activite_registre_metiers,
			etat_administratif, enseigne, denomination_usuelle, caractere_employeur, insert) values 
		($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,
		 $12,$13,$14,$15,$16,$17,$18,$19,$20,
		 $21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31);`
	for _, sirene := range data {
		d := geoSireneData(sirene)
		batch.Queue(sql, d...)
	}

	batch.Queue(`
	  insert into etablissement
		( siret, siren, siege, creation, complement_adresse, numero_voie,
			indice_repetition, type_voie, voie, commune, commune_etranger, distribution_speciale,
			code_commune, code_cedex, cedex, code_pays_etranger, pays_etranger, code_postal,
			departement, code_activite, nomen_activite, latitude, longitude, 
		  tranche_effectif, annee_effectif, code_activite_registre_metiers,
			etat_administratif, enseigne, denomination_usuelle, caractere_employeur, perimetre) select
			s.siret,
			s.siren,
			s.siege,
			s.creation,
			s.complement_adresse,
			s.numero_voie,
			s.indice_repetition,
			s.type_voie,
			s.voie,
			s.commune,
			s.commune_etranger,
			s.distribution_speciale,
			s.code_commune,
			s.code_cedex,
			s.cedex,
			s.code_pays_etranger,
			s.pays_etranger,
			s.code_postal,
			s.departement,
			s.code_activite,
			s.nomen_activite,
			s.latitude,
			s.longitude,
			s.tranche_effectif,
			s.annee_effectif,
			s.code_activite_registre_metiers,
			s.etat_administratif,
			s.enseigne,
			s.denomination_usuelle,
			s.caractere_employeur,
			false
		from tmp_geosirene s
		where s.insert
	`)
	results := tx.SendBatch(ctx, batch)
	err = results.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	err = tx.Commit(context.Background())
	return err
}

func insertSireneUL(ctx context.Context, cancelCtx context.CancelFunc, wg *sync.WaitGroup) {
	sireneUL := goSirene.SireneULParser(ctx, viper.GetString("sireneULPath"))
	count := 0
	var batch []goSirene.SireneUL
	for sirene := range sireneUL {
		count++
		if count == 100000 {
			count = 0
			err := sqlSireneUL(ctx, batch)
			if err != nil {
				cancelCtx()
				fmt.Println(err.Error())
				continue
			}
			batch = []goSirene.SireneUL{}
			log.Println("insert SireneUL")
		}
		if sirene.Error() == nil {
			batch = append(batch, sirene)
		}
	}
	err := sqlSireneUL(ctx, batch)
	if err != nil {
		cancelCtx()
		fmt.Println(err.Error())
	}
	wg.Done()
}

func sqlSireneUL(ctx context.Context, data []goSirene.SireneUL) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	tx.Exec(ctx, `
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
	identifiant_association text, 
	tranche_effectif text,
	annee_effectif text, 
	categorie text, 
	annee_categorie text, 
	etat_administratif text, 
	economie_sociale_solidaire boolean,
	caractere_employeur boolean,
  insert bool)
	on commit drop;`)

	batch := &pgx.Batch{}
	sql := `
	  insert into tmp_sirene_ul (siren,
		siret_siege, 
		raison_sociale, 
		prenom1, 
		prenom2, 
		prenom3, 
		prenom4, 
		nom,
		nom_usage, 
		statut_juridique, 
		creation, 
		sigle, 
		identifiant_association, 
		tranche_effectif,
		annee_effectif, 
		categorie, 
		annee_categorie, 
		etat_administratif, 
		economie_sociale_solidaire,
		caractere_employeur) values 
		($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,
		 $12,$13,$14,$15,$16,$17,$18,$19,$20);`
	for _, sirene := range data {
		d := sireneULdata(sirene)
		batch.Queue(sql, d...)
	}
	batch.Queue(`
	  insert into entreprise
		(siren, siret_siege, raison_sociale, prenom1, prenom2, prenom3,
			prenom4, nom, nom_usage, statut_juridique, creation, sigle,
			identifiant_association, tranche_effectif, annee_effectif,
			categorie, annee_categorie, etat_administratif,
		  economie_sociale_solidaire, caractere_employeur) select
		ul.siren,
		ul.siret_siege,
		ul.raison_sociale,
		ul.prenom1,
		ul.prenom2,
		ul.prenom3,
		ul.prenom4,
		ul.nom,
		ul.nom_usage,
		ul.statut_juridique,
		ul.creation,
		ul.sigle,
		ul.identifiant_association,
		ul.tranche_effectif,
		ul.annee_effectif,
		ul.categorie,
		ul.annee_categorie,
		ul.etat_administratif,
		ul.economie_sociale_solidaire,
		ul.caractere_employeur
		from tmp_sirene_ul ul
	`)
	results := tx.SendBatch(ctx, batch)
	err = results.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	err = tx.Commit(context.Background())
	return err
}

func sireneULdata(s goSirene.SireneUL) []interface{} {
	return []interface{}{
		null(s.Siren),
		null(s.NicSiegeUniteLegale),
		null(s.RaisonSociale()),
		null(s.Prenom1UniteLegale),
		null(s.Prenom2UniteLegale),
		null(s.Prenom3UniteLegale),
		null(s.Prenom4UniteLegale),
		null(s.NomUniteLegale),
		null(s.NomUsageUniteLegale),
		null(s.CategorieJuridiqueUniteLegale),
		s.DateCreationUniteLegale,
		null(s.SigleUniteLegale),
		null(s.IdentifiantAssociationUniteLegale),
		null(s.TrancheEffectifsUniteLegale),
		null(s.AnneeEffectifsUniteLegale),
		null(s.CategorieEntreprise),
		null(s.AnneeCategorieEntreprise),
		null(s.EtatAdministratifUniteLegale),
		s.EconomieSocialeSolidaireUniteLegale,
		s.CaractereEmployeurUniteLegale,
	}
}

func geoSireneData(s goSirene.GeoSirene) []interface{} {
	return []interface{}{
		null(s.Siret),
		null(s.Siren),
		s.EtablissementSiege,
		s.DateCreationEtablissement,
		null(s.ComplementAdresseEtablissement),
		null(s.NumeroVoieEtablissement),
		null(s.IndiceRepetitionEtablissement),
		null(s.TypeVoieEtablissement),
		null(s.LibelleVoieEtablissement),
		null(s.LibelleCommuneEtablissement),
		null(s.LibelleCommuneEtrangerEtablissement),
		null(s.DistributionSpecialeEtablissement),
		null(s.CodeCommuneEtablissement),
		null(s.CodeCedexEtablissement),
		null(s.LibelleCedexEtablissement),
		null(s.CodePaysEtrangerEtablissement),
		null(s.LibellePaysEtrangerEtablissement),
		null(s.CodePostalEtablissement),
		null(s.CodeDepartement()),
		null(strings.Replace(s.ActivitePrincipaleEtablissement, ".", "", -1)),
		null(s.NomenclatureActivitePrincipaleEtablissement),
		s.Latitude,
		s.Longitude,
		null(s.TrancheEffectifsEtablissement),
		null(s.AnneeEffectifsEtablissement),
		null(s.ActivitePrincipaleRegistreMetiersEtablissement),
		null(s.EtatAdministratifEtablissement),
		null(s.Enseigne1Etablissement),
		null(s.DenominationUsuelleEtablissement),
		s.CaractereEmployeurEtablissement,
	}
}

func null(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func truncateSirens(tx pgx.Tx) error {
	_, err := tx.Exec(context.Background(), "truncate table etablissement")
	if err != nil {
		return err
	}
	_, err = tx.Exec(context.Background(), "truncate table entreprise")
	return err
}

type Siret string

func (s Siret) Less(than btree.Item) bool {
	return string(s) > fmt.Sprint(than)
}

func (s Siret) Siren() Siret {
	return s[0:9]
}
