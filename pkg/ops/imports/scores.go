// Package imports contient le code lié aux opérations d'administration dans datapi
package imports

import (
	"context"
	_ "embed"
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
	Siren     string             `json:"siren"`
	Score     float64            `json:"probability"`
	Alert     string             `json:"alert"`
	MacroExpl map[string]float64 `json:"macroExpl"`
	MicroExpl map[string]float64 `json:"microExpl"`
}

func (s score) Values(batch string, algo string) []any {
	return []any{
		s.Siren,
		s.Score,
		s.Alert,
		batch,
		algo,
		s.MicroExpl,
		s.MacroExpl,
		toLibelle(batch),
	}
}

type CopyFromScore struct {
	scores      *[]score
	batchNumber string
	algo        string
	current     *int
	err         *error
	ready       bool
}

func parseScores(filename string) ([]score, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, errors.New("open file: " + err.Error())
	}
	raw, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.New("read file: " + err.Error())
	}
	err = file.Close()
	if err != nil {
		return nil, errors.New("close file: " + err.Error())
	}
	var scores []score
	err = json.Unmarshal(raw, &scores)

	return scores, err
}

func (c CopyFromScore) isReady() bool {
	return c.ready
}

func (c *CopyFromScore) init(filename string, batchNumber string, algo string) error {
	c.current = new(int)
	c.batchNumber = batchNumber
	c.algo = algo

	scores, err := parseScores(filename)
	if err != nil {
		c.err = &err
		return err
	}
	c.scores = &scores
	c.ready = true

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
		return (*c.scores)[*c.current].Values(c.batchNumber, c.algo), nil
	}

	return nil, CopyFromSourceDepleted{"CopyFromScore"}
}

func (c CopyFromScore) Err() error {
	if c.err == nil {
		return nil
	}
	return *c.err
}

//go:embed sql/scoreUpdateSiretWithSiren.sql
var sqlScoreUpdateSiretWithSiren string

func importListe(ctx context.Context, batchNumber string, algo string) error {
	filename := viper.GetString("source.listPath")
	copyFromScore := CopyFromScore{}
	err := copyFromScore.init(filename, batchNumber, algo)
	if err != nil {
		return err
	}

	tx, err := db.Get().Begin(context.Background())
	if err != nil {
		return utils.NewJSONerror(http.StatusBadRequest, "begin TX: "+err.Error())
	}

	_, err = tx.Exec(context.Background(), `delete from liste where batch=$1`, batchNumber)
	if err != nil {
		return errors.New("remove liste: " + err.Error())
	}
	_, err = tx.Exec(context.Background(), `delete from score where batch=$1;`, batchNumber)
	if err != nil {
		return errors.New("remove scores: " + err.Error())
	}

	columns := []string{
		"siren",
		"score",
		"alert",
		"batch",
		"algo",
		"micro_expl",
		"macro_expl",
		"libelle_liste",
	}
	_, err = tx.CopyFrom(ctx, pgx.Identifier{"score"}, columns, copyFromScore)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	_, err = tx.Exec(ctx, sqlScoreUpdateSiretWithSiren, batchNumber)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	_, err = tx.Exec(ctx, "insert into liste (batch, libelle, algo) values ($1, $2, $3)", batchNumber, toLibelle(batchNumber))
	err = tx.Commit(context.Background())
	if err != nil {
		return errors.New("commit: " + err.Error())
	}
	return nil
}
