package imports

import (
	"context"
	"datapi/pkg/db"
	"github.com/jackc/pgx/v5"
	"github.com/signaux-faibles/goSirene"
	"github.com/spf13/viper"
	"log"
	"strings"
)

// InsertSireneUL insère des trucs rapport aux sirene mais je sais pas trop quoi
func InsertSireneUL(ctx context.Context) error {
	sireneULParser := goSirene.SireneULParser(ctx, viper.GetString("sireneULPath"))
	initialCount := 0
	copyFromSireneUL := CopyFromSireneUL{
		SireneULParser: sireneULParser,
		Count:          &initialCount,
		Current:        &goSirene.SireneUL{},
	}

	return copySireneUL(ctx, copyFromSireneUL)
}

type CopyFromSireneUL struct {
	SireneULParser chan goSirene.SireneUL
	Current        *goSirene.SireneUL
	Count          *int
}

func (c CopyFromSireneUL) Increment() {
	*c.Count = *c.Count + 1
}

func (c CopyFromSireneUL) Next() bool {
	var ok bool
	select {
	case *c.Current, ok = <-c.SireneULParser:
		if ok {
			if c.Current.Error() != nil {
				return c.Next()
			}
			c.Increment()
			if *c.Count%100000 == 0 {
				log.Printf("%d sireneUL object copied\n", *c.Count)
			}
		}
		return ok
	}
}

func (c CopyFromSireneUL) Err() error { return nil }
func (c CopyFromSireneUL) Values() ([]interface{}, error) {
	return sireneULdata(*c.Current), nil
}

func copySireneUL(ctx context.Context, copyFromSource pgx.CopyFromSource) error {
	conn := db.Get()
	headers := []string{"siren", "siret_siege", "raison_sociale", "prenom1", "prenom2", "prenom3",
		"prenom4", "nom", "nom_usage", "statut_juridique", "creation", "sigle",
		"identifiant_association", "tranche_effectif", "annee_effectif",
		"categorie", "annee_categorie", "etat_administratif",
		"economie_sociale_solidaire", "caractere_employeur",
		"code_activite", "nomen_activite"}
	identifier := pgx.Identifier{"entreprise"}
	_, err := conn.CopyFrom(ctx, identifier, headers, copyFromSource)
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
		boolToText(s.CaractereEmployeurUniteLegale),
		null(strings.Replace(s.ActivitePrincipaleUniteLegale, ".", "", 1)),
		s.NomenclatureActivitePrincipaleUniteLegale,
	}
}

func boolToText(b bool) string {
	if b {
		return "true"
	} else {
		return "false"
	}
}

func null(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func TruncateEntreprise(ctx context.Context) error {
	_, err := db.Get().Exec(ctx, "truncate table entreprise")
	return err
}

func CreateEntrepriseIndex(ctx context.Context) error {
	conn := db.Get()
	sql := `create index idx_entreprise_siren on entreprise (siren);`
	_, err := conn.Exec(ctx, sql)
	return err
}

func DropEntrepriseIndex(ctx context.Context) error {
	conn := db.Get()
	sql := `drop index if exists idx_entreprise_siren;`
	_, err := conn.Exec(ctx, sql)
	return err
}
