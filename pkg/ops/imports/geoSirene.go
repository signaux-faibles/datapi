package imports

import (
	"context"
	"datapi/pkg/db"
	"github.com/jackc/pgx/v5"
	"github.com/signaux-faibles/goSirene"
	"github.com/spf13/viper"
	"log"
	"os"
	"strings"
)

func CreateEtablissementIndex(ctx context.Context) error {
	conn := db.Get()
	sql := `create index idx_etablissement_siret on etablissement (siret) where version = 0;
	create index idx_etablissement_siren on etablissement (siren) where version = 0;
	create index idx_etablissement_departement on etablissement (departement) where version = 0;
	create index idx_etablissement_naf on etablissement (code_activite) where version = 0;`
	_, err := conn.Exec(ctx, sql)
	return err
}

func DropEtablissementIndex(ctx context.Context) error {
	conn := db.Get()
	sql := `drop index if exists idx_etablissement_siret;
	drop index if exists idx_etablissement_siren;
	drop index if exists idx_etablissement_departement;
	drop index if exists idx_etablissement_naf;`
	_, err := conn.Exec(ctx, sql)
	return err
}

// InsertGeoSirene insère les informations géographiques des établissements dans la base
func InsertGeoSirene(ctx context.Context) error {
	file, err := os.Open(viper.GetString("geoSirenePath"))
	if err != nil {
		return err
	}
	geoSireneParser := goSirene.GeoSireneParser(ctx, file)
	copyFromGeoSirene := CopyFromGeoSirene{
		GeoSireneParser: geoSireneParser,
		Current:         new(goSirene.GeoSirene),
		Count:           new(int),
	}
	return copyGeoSirene(ctx, copyFromGeoSirene)
}

type CopyFromGeoSirene struct {
	GeoSireneParser chan goSirene.GeoSirene
	Current         *goSirene.GeoSirene
	Count           *int
}

func (c CopyFromGeoSirene) Increment() {
	*c.Count = *c.Count + 1
}

func (c CopyFromGeoSirene) Next() bool {
	var ok bool
	select {
	case *c.Current, ok = <-c.GeoSireneParser:
		if ok {
			if c.Current.Error() != nil {
				return c.Next()
			}
			c.Increment()
			if *c.Count%100000 == 0 {
				log.Printf("%d geoSirene object copied\n", *c.Count)
			}
		}
		return ok
	}
}

func (c CopyFromGeoSirene) Err() error { return nil }
func (c CopyFromGeoSirene) Values() ([]interface{}, error) {
	return geoSireneData(*c.Current), nil
}

func copyGeoSirene(ctx context.Context, copyFromSource pgx.CopyFromSource) error {
	conn := db.Get()
	headers := []string{"siret", "siren", "siege", "creation", "complement_adresse", "numero_voie",
		"indice_repetition", "type_voie", "voie", "commune", "commune_etranger", "distribution_speciale",
		"code_commune", "code_cedex", "cedex", "code_pays_etranger", "pays_etranger", "code_postal",
		"departement", "code_activite", "nomen_activite", "latitude", "longitude",
		"tranche_effectif", "annee_effectif", "code_activite_registre_metiers",
		"etat_administratif", "enseigne", "denomination_usuelle", "caractere_employeur"}
	identifier := pgx.Identifier{"etablissement"}

	_, err := conn.CopyFrom(ctx, identifier, headers, copyFromSource)
	return err
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

func TruncateEtablissement() error {
	_, err := db.Get().Exec(context.Background(), "truncate table etablissement")
	return err
}
