// Package imports contient le code lié aux opérations d'administration dans datapi
package imports

import (
	"context"
	"github.com/spf13/viper"
	"log/slog"
	"net/http"

	"datapi/pkg/utils"
)

func toLibelle(batchNumber string) string {
	months := map[string]string{
		"01": "Janvier",
		"02": "Février",
		"03": "Mars",
		"04": "Avril",
		"05": "Mai",
		"06": "Juin",
		"07": "Juillet",
		"08": "Août",
		"09": "Septembre",
		"10": "Octobre",
		"11": "Novembre",
		"12": "Décembre",
	}
	year := "20" + batchNumber[0:2]
	month := batchNumber[2:4]
	return months[month] + " " + year
}

func importUnitesLegales(ctx context.Context) error {
	if viper.GetString("source.sireneULPath") == "" || viper.GetString("source.geoSirenePath") == "" {
		return utils.NewJSONerror(http.StatusConflict, "not supported, missing parameters in server configuration")
	}
	slog.Info("Truncate entreprise table", slog.String("status", "start"))

	err := TruncateEntreprise(ctx)
	if err != nil {
		return err
	}
	slog.Info("Truncate entreprise table", slog.String("status", "end"))

	err = DropEntrepriseIndex(ctx)
	if err != nil {
		return err
	}

	err = InsertSireneUL(ctx)
	if err != nil {
		return err
	}

	return CreateEntrepriseIndex(ctx)
}

func importStockEtablissement(ctx context.Context) error {
	if viper.GetString("source.geoSirenePath") == "" {
		return utils.NewJSONerror(http.StatusConflict, "not supported, missing geoSirenePath parameters in server configuration")
	}
	slog.Info("Truncate etablissement table")
	err := TruncateEtablissement()
	if err != nil {
		return err
	}

	slog.Info("Drop etablissement table indexes")
	err = DropEtablissementIndex(ctx)
	if err != nil {
		return err
	}
	err = InsertGeoSirene(ctx)
	if err != nil {
		return err
	}
	slog.Info("Create etablissement table indexes")
	return CreateEtablissementIndex(ctx)
}
