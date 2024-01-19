// Package imports contient le code lié aux opérations d'administration dans datapi
package imports

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"os"

	"datapi/pkg/db"
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

type score struct {
	Siren                 string             `json:"siren"`
	Score                 float64            `json:"score"`
	Diff                  float64            `json:"diff"`
	Alert                 string             `json:"alert"`
	Periode               string             `json:"periode"`
	Batch                 string             `json:"batch"`
	Algo                  string             `json:"algo"`
	MacroExpl             map[string]float64 `json:"macroExpl"`
	MicroExpl             map[string]float64 `json:"microExpl"`
	Redressements         []string           `json:"redressements"`
	AlertPreRedressements string             `json:"alertPreRedressements"`
}

func (s score) Values(batch string, algo string) []any {
	//    "siren",
	//    "libelle_liste",
	//    "batch",
	//    "algo",
	//    "periode",
	//    "score",
	//    "alert",
	//    "micro_expl",
	//    "macro_expl",
	//
	return []any{
		s.Siren,
		toLibelle(s.Batch),
        s.Algo,
        date
	}

}

type CopyFromScore struct {
	scores  *[]score
	current *int
	err     *error
	ready   *bool
}

func (c CopyFromScore) isReady() bool {
	return c.ready != nil
}

func (c CopyFromScore) init(filename string) error {
	zero := 0
	c.current = &zero

	file, err := os.Open(filename)
	if err != nil {
		return errors.New("open file: " + err.Error())
	}
	raw, err := io.ReadAll(file)
	file.Close()
	if err != nil {
		return errors.New("read file: " + err.Error())
	}
	err = json.Unmarshal(raw, c.scores)
	if err != nil {
		if err != nil {
			return errors.New("unmarshall JSON : " + err.Error())
		}
	}

	c.ready = new(bool)
	return nil
}

func (c CopyFromScore) Next() bool {
	if !c.isReady() {
		return false
	}
	*c.current = *c.current + 1
	if *c.current < len(*c.scores) {
		return true
	}
	return false
}

func (c CopyFromScore) Values() ([]any, error) {
	if !c.isReady() {
		return nil, CopyFromSourceNotReady{"CopyFromScore"}
	}

	if *c.current < len(*c.scores) {
		return (*c.scores)[*c.current].Values(), nil
	}

	return nil, CopyFromSourceDepleted{"CopyFromScore"}
}

func (c CopyFromScore) Err() error {
	if c.err == nil {
		return nil
	}
	return *c.err
}

func importListe(ctx context.Context, batchNumber string, algo string) error {
	filename := viper.GetString("source.listPath")
	copyFromScore := CopyFromScore{}
	err := copyFromScore.init(filename)
	if err != nil {
		return err
	}

	tx, err := db.Get().Begin(context.Background())
	if err != nil {
		return utils.NewJSONerror(http.StatusBadRequest, "begin TX: "+err.Error())
	}

	_, err = tx.Exec(context.Background(), `truncate table score;`)
	if err != nil {
		return errors.New("truncate score: " + err.Error())
	}

	columns := []string{
		"date_add",
		"siret",
		"siren",
		"libelle_liste",
		"batch",
		"algo",
		"periode",
		"score",
		"alert",
		"micro_expl",
		"macro_expl",
	}
	tx.CopyFrom(ctx, pgx.Identifier{"score"}, columns, copyFromScore)
	err = tx.Commit(context.Background())
	if err != nil {
		return errors.New("commit: " + err.Error())
	}
	return nil
}
