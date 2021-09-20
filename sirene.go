package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"github.com/signaux-faibles/goSirene"
	"github.com/spf13/viper"
)

func sireneImportHandler(c *gin.Context) {
	if viper.GetString("sireneULPath") == "" || viper.GetString("geoSirenePath") == "" {
		c.AbortWithStatusJSON(409, "not supported, missing parameters in server configuration")
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	ctx, cancelCtx := context.WithCancel(context.Background())
	go upsertSireneUL(ctx, cancelCtx, &wg)
	go upsertGeoSirene(ctx, cancelCtx, &wg)

	wg.Wait()
}

func upsertGeoSirene(ctx context.Context, cancelCtx context.CancelFunc, wg *sync.WaitGroup) {
	file, err := os.Open(viper.GetString("geoSirenePath"))
	if err != nil {
		cancelCtx()
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
		caractere_employeur bool)
	on commit drop;`)

	batch := &pgx.Batch{}
	sql := `
	  insert into tmp_geosirene (siret, siren, siege, creation, complement_adresse, numero_voie,
			indice_repetition, type_voie, voie, commune, commune_etranger, distribution_speciale,
			code_commune, code_cedex, cedex, code_pays_etranger, pays_etranger, code_postal,
			departement, code_activite, nomen_activite, latitude, longitude, 
		  tranche_effectif, annee_effectif, code_activite_registre_metiers,
			etat_administratif, enseigne, denomination_usuelle, caractere_employeur) values 
		($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,
		 $12,$13,$14,$15,$16,$17,$18,$19,$20,
		 $21,$22,$23,$24,$25,$26,$27,$28,$29,$30);`
	for _, sirene := range data {
		d := geoSireneData(sirene)
		batch.Queue(sql, d...)
	}
	batch.Queue(`
	  update etablissement e set 
		siren = t.siren,
		siege = t.siege,
		creation = t.creation,
		complement_adresse = t.complement_adresse,
		numero_voie = t.numero_voie,
		indice_repetition = t.indice_repetition,
		type_voie = t.type_voie,
		voie = t.voie,
		commune = t.commune,
		commune_etranger = t.commune_etranger,
		distribution_speciale = t.distribution_speciale,
		code_commune = t.code_commune,
		code_cedex = t.code_cedex,
		cedex = t.cedex,
		code_pays_etranger = t.code_pays_etranger,
		pays_etranger = t.pays_etranger,
		code_postal = t.code_postal,
		departement = t.departement,
		code_activite = t.code_activite,
		nomen_activite = t.nomen_activite,
		latitude = t.latitude,
		longitude = t.longitude,
	  tranche_effectif = t.tranche_effectif,
		annee_effectif = t.annee_effectif,
		code_activite_registre_metiers = t.code_activite_registre_metiers,
		etat_administratif = t.etat_administratif,
		enseigne = t.enseigne,
		denomination_usuelle = t.denomination_usuelle,
		caractere_employeur = t.caractere_employeur
		from tmp_geosirene t where t.siret = e.siret and e.version = 0;
	`)
	batch.Queue(`
	  insert into etablissement
		( siret, siren, siege, creation, complement_adresse, numero_voie,
			indice_repetition, type_voie, voie, commune, commune_etranger, distribution_speciale,
			code_commune, code_cedex, cedex, code_pays_etranger, pays_etranger, code_postal,
			departement, code_activite, nomen_activite, latitude, longitude, 
		  tranche_effectif, annee_effectif, code_activite_registre_metiers,
			etat_administratif, enseigne, denomination_usuelle, caractere_employeur) select
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
			s.caractere_employeur
		from tmp_geosirene s
		left join entreprise e on e.siren = s.siren and e.version = 0
		where e.siren is null
	`)
	results := tx.SendBatch(ctx, batch)
	err = results.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	err = tx.Commit(context.Background())
	return err
}

func upsertSireneUL(ctx context.Context, cancelCtx context.CancelFunc, wg *sync.WaitGroup) {
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
			log.Println("upsert SireneUL")
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
	caractere_employeur boolean)
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
	  update entreprise e set 
		siret_siege = t.siret_siege,
		raison_sociale = t.raison_sociale,
		prenom1 = t.prenom1,
		prenom2 = t.prenom2,
		prenom3 = t.prenom3,
		prenom4 = t.prenom4,
		nom = t.nom,
		nom_usage = t.nom_usage,
		statut_juridique = t.statut_juridique,
		creation = t.creation,
		sigle = t.sigle,
		identifiant_association = t.identifiant_association,
		tranche_effectif = t.tranche_effectif,
		annee_effectif = t.annee_effectif,
		categorie = t.categorie,
		annee_categorie = t.annee_categorie,
		etat_administratif = t.etat_administratif,
		economie_sociale_solidaire = t.economie_sociale_solidaire
		from tmp_sirene_ul t where t.siren = e.siren and e.version = 0;
	`)
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
		left join entreprise e on e.siren = ul.siren and e.version = 0
		where e.siren is null
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
		s.Siren,
		s.NicSiegeUniteLegale,
		s.RaisonSociale(),
		s.Prenom1UniteLegale,
		s.Prenom2UniteLegale,
		s.Prenom3UniteLegale,
		s.Prenom4UniteLegale,
		s.NomUniteLegale,
		s.NomUsageUniteLegale,
		s.CategorieJuridiqueUniteLegale,
		s.DateCreationUniteLegale,
		s.SigleUniteLegale,
		s.IdentifiantAssociationUniteLegale,
		s.TrancheEffectifsUniteLegale,
		s.AnneeEffectifsUniteLegale,
		s.CategorieEntreprise,
		s.AnneeCategorieEntreprise,
		s.EtatAdministratifUniteLegale,
		s.EconomieSocialeSolidaireUniteLegale,
		s.CaractereEmployeurUniteLegale,
	}
}

func geoSireneData(s goSirene.GeoSirene) []interface{} {
	return []interface{}{
		s.Siret,
		s.Siren,
		s.EtablissementSiege,
		s.DateCreationEtablissement,
		s.ComplementAdresseEtablissement,
		s.NumeroVoieEtablissement,
		s.IndiceRepetitionEtablissement,
		s.TypeVoieEtablissement,
		s.LibelleVoieEtablissement,
		s.LibelleCommuneEtablissement,
		s.LibelleCommuneEtrangerEtablissement,
		s.DistributionSpecialeEtablissement,
		s.CodeCommuneEtablissement,
		s.CodeCedexEtablissement,
		s.LibelleCedexEtablissement,
		s.CodePaysEtrangerEtablissement,
		s.LibellePaysEtrangerEtablissement,
		s.CodePostalEtablissement,
		s.CodeDepartement(),
		s.ActivitePrincipaleEtablissement,
		s.NomenclatureActivitePrincipaleEtablissement,
		s.Latitude,
		s.Longitude,
		s.TrancheEffectifsEtablissement,
		s.AnneeEffectifsEtablissement,
		s.ActivitePrincipaleRegistreMetiersEtablissement,
		s.EtatAdministratifEtablissement,
		s.Enseigne1Etablissement,
		s.DenominationUsuelleEtablissement,
		s.CaractereEmployeurEtablissement,
	}
}
