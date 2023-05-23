package imports

import (
	"context"
	"encoding/csv"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/signaux-faibles/datapi/src/db"
	"github.com/signaux-faibles/datapi/src/utils"
	"github.com/spf13/viper"
	"os"
	"strconv"
	"strings"
	"time"
)

type NullError struct{}

type BCE struct {
	Siren                           string
	DateClotureExercice             time.Time
	ChiffreDAffaires                *int64
	MargeBrute                      *int64
	EBE                             *int64
	EBIT                            *int64
	ResultatNet                     *int64
	TauxDEndettement                *float64
	RatioDeLiquidite                *float64
	RatioDeVetuste                  *float64
	AutonomieFinanciere             *float64
	PoidsBFRExploitationSurCA       *float64
	CouvertureDesInterets           *float64
	CAFsurCA                        *float64
	CapaciteDeRemboursement         *float64
	MargeEBE                        *float64
	ResultatCourantAvantImpotsSurCA *float64
	PoidsBFRExploitationSurCAJours  *float64
	RotationDesStocksJours          *float64
	CreditClientsJours              *float64
	CreditFournisseursJours         *float64
	TypeBilan                       string
}

func (bce BCE) fields() []interface{} {
	return []interface{}{
		bce.Siren,
		bce.DateClotureExercice,
		bce.ChiffreDAffaires,
		bce.MargeBrute,
		bce.EBE,
		bce.EBIT,
		bce.ResultatNet,
		bce.TauxDEndettement,
		bce.RatioDeLiquidite,
		bce.RatioDeVetuste,
		bce.AutonomieFinanciere,
		bce.PoidsBFRExploitationSurCA,
		bce.CouvertureDesInterets,
		bce.CAFsurCA,
		bce.CapaciteDeRemboursement,
		bce.MargeEBE,
		bce.ResultatCourantAvantImpotsSurCA,
		bce.PoidsBFRExploitationSurCAJours,
		bce.RotationDesStocksJours,
		bce.CreditClientsJours,
		bce.CreditFournisseursJours,
		bce.TypeBilan,
	}
}

func importBCEHandler(c *gin.Context) {
	sourcePath := viper.GetString("bceSourcePath")
	db := db.Get()
	err := importBCE(c, sourcePath, db)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	c.JSON(200, "ok")
}
func parseBCE(input []string) (bce BCE) {
	var err error

	bce.DateClotureExercice, err = time.Parse("20060102", input[1])
	if err != nil {
		return BCE{}
	}
	bce.ChiffreDAffaires, err = parseInt(input[2])
	if err != nil {
		return BCE{}
	}
	bce.MargeBrute, err = parseInt(input[3])
	if err != nil {
		return BCE{}
	}
	bce.EBE, err = parseInt(input[4])
	if err != nil {
		return BCE{}
	}
	bce.EBIT, err = parseInt(input[5])
	if err != nil {
		return BCE{}
	}
	bce.ResultatNet, err = parseInt(input[6])
	if err != nil {
		return BCE{}
	}
	bce.TauxDEndettement, err = parseFloat(input[7])
	if err != nil {
		return BCE{}
	}
	bce.RatioDeLiquidite, err = parseFloat(input[8])
	if err != nil {
		return BCE{}
	}
	bce.RatioDeVetuste, err = parseFloat(input[9])
	if err != nil {
		return BCE{}
	}
	bce.AutonomieFinanciere, err = parseFloat(input[10])
	if err != nil {
		return BCE{}
	}
	bce.PoidsBFRExploitationSurCA, err = parseFloat(input[11])
	if err != nil {
		return BCE{}
	}
	bce.CouvertureDesInterets, err = parseFloat(input[12])
	if err != nil {
		return BCE{}
	}
	bce.CAFsurCA, err = parseFloat(input[13])
	if err != nil {
		return BCE{}
	}
	bce.CapaciteDeRemboursement, err = parseFloat(input[14])
	if err != nil {
		return BCE{}
	}
	bce.MargeEBE, err = parseFloat(input[15])
	if err != nil {
		return BCE{}
	}
	bce.ResultatCourantAvantImpotsSurCA, err = parseFloat(input[16])
	if err != nil {
		return BCE{}
	}
	bce.PoidsBFRExploitationSurCAJours, err = parseFloat(input[17])
	if err != nil {
		return BCE{}
	}
	bce.RotationDesStocksJours, err = parseFloat(input[18])
	if err != nil {
		return BCE{}
	}
	bce.CreditClientsJours, err = parseFloat(input[19])
	if err != nil {
		return BCE{}
	}
	bce.CreditFournisseursJours, err = parseFloat(input[20])
	if err != nil {
		return BCE{}
	}
	bce.Siren = input[0]
	bce.TypeBilan = input[21]
	return bce
}

func parseFloat(s string) (*float64, error) {
	if s == "" {
		return nil, nil
	}
	f, err := strconv.ParseFloat(strings.Replace(s, ",", ".", 1), 64)
	return &f, err
}

func parseInt(s string) (*int64, error) {
	if s == "" {
		return nil, nil
	}
	i, err := strconv.ParseInt(s, 10, 64)
	return &i, err
}

func populateDiane(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, "truncate table entreprise_diane;")
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, "insert into entreprise_diane (siren, arrete_bilan_diane, exercice_diane, chiffre_affaire, resultat_expl, excedent_brut_d_exploitation, benefice_ou_perte)"+
		"select siren,"+
		"  date_cloture_exercice,"+
		"  extract(year from date_cloture_exercice - '6 month'::interval),"+
		"  round(first(chiffre_d_affaires order by type_bilan != 'K')/1000),"+
		"  round(first(ebit order by type_bilan != 'K')/1000),"+
		"  round(first(ebe order by type_bilan != 'K')/1000),"+
		"  round(first(resultat_net order by type_bilan != 'K')/1000)"+
		"  from entreprise_bce"+
		"  group by siren, date_cloture_exercice"+
		"  order by siren, date_cloture_exercice desc")
	return err
}

func truncateBCE(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, "truncate table entreprise_bce;")
	return err
}

func importBCE(ctx context.Context, path string, db *pgxpool.Pool) error {
	bceSourceFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer bceSourceFile.Close()

	bceReader := csv.NewReader(bceSourceFile)
	bceReader.Comma = ';'
	data, err := bceReader.ReadAll()

	bces := utils.Convert(data, parseBCE)
	rows := utils.Convert(bces, BCE.fields)
	err = truncateBCE(ctx, db)
	if err != nil {
		return err
	}

	_, err = db.CopyFrom(ctx, pgx.Identifier{"entreprise_bce"}, bceColums, pgx.CopyFromRows(rows))
	if err != nil {
		return err
	}
	err = populateDiane(ctx, db)
	return err
}

var bceColums = []string{
	"siren",
	"date_cloture_exercice",
	"chiffre_d_affaires",
	"marge_brute",
	"ebe",
	"ebit",
	"resultat_net",
	"taux_d_endettement",
	"ratio_de_liquidite",
	"ratio_de_vetuste",
	"autonomie_financiere",
	"poids_bfr_exploitation_sur_ca",
	"couverture_des_interets",
	"caf_sur_ca",
	"capacite_de_remboursement",
	"marge_ebe",
	"resultat_courant_avant_impots_sur_ca",
	"poids_bfr_exploitation_sur_ca_jours",
	"rotation_des_stocks_jours",
	"credit_clients_jours",
	"credit_fournisseurs_jours",
	"type_bilan",
}
