package imports

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"datapi/pkg/db"
	"datapi/pkg/utils"
)

type score struct {
	Siret         string        `json:"-"`
	Siren         string        `json:"siret"`
	Libelle       string        `json:"-"`
	ID            bson.ObjectId `json:"-"`
	Score         float64       `json:"score"`
	Diff          float64       `json:"diff"`
	Timestamp     time.Time     `json:"-"`
	Alert         string        `json:"alert"`
	Periode       string        `json:"periode"`
	Batch         string        `json:"batch"`
	Algo          string        `json:"algo"`
	ExplSelection struct {
		SelectConcerning [][]string `json:"select_concerning"`
		SelectReassuring [][]string `json:"select_reassuring"`
	} `json:"expl_selection"`
	MacroExpl             map[string]float64 `json:"macro_expl"`
	MicroExpl             map[string]float64 `json:"micro_expl"`
	MacroRadar            map[string]float64 `json:"macro_radar"`
	AlertPreRedressements string             `json:"alert_pre_redressements"`
}

type scoreFile struct {
	Siren         string  `json:"siren"`
	Score         float64 `json:"score"`
	Diff          float64 `json:"diff"`
	Alert         string  `json:"alert"`
	Periode       string  `json:"periode"`
	Batch         string  `json:"batch"`
	Algo          string  `json:"algo"`
	ExplSelection struct {
		SelectConcerning [][]string `json:"selectConcerning"`
		SelectReassuring [][]string `json:"selectReassuring"`
	} `json:"explSelection"`
	MacroRadar            map[string]float64 `json:"macroRadar"`
	Redressements         []string           `json:"redressements"`
	AlertPreRedressements string             `json:"alertPreRedressements"`
}

func importListe(batchNumber string, algo string) error {
	filename := viper.GetString("source.listPath")
	file, err := os.Open(filename)
	if err != nil {
		return errors.New("open file: " + err.Error())
	}
	raw, err := io.ReadAll(file)
	file.Close()
	if err != nil {
		return errors.New("read file: " + err.Error())
	}
	var scores []scoreFile
	err = json.Unmarshal(raw, &scores)
	if err != nil {
		if err != nil {
			return errors.New("unmarshall JSON : " + err.Error())
		}
	}
	now := time.Now()
	tx, err := db.Get().Begin(context.Background())
	if err != nil {
		return utils.NewJSONerror(http.StatusBadRequest, "begin TX: "+err.Error())
	}

	_, err = tx.Exec(context.Background(), `drop table if exists tmp_score;
        create table tmp_score (
		siren text,
		score real,
		libelle_liste text,
		diff real,
		alert text,
		batch text,
		algo text,
		expl_selection_concerning jsonb default '{}',
		expl_selection_reassuring jsonb default '{}',
		macro_radar jsonb default '{}',
		alert_pre_redressements text,
		redressements text[] default '{}'
	);`)
	if err != nil {
		return errors.New("create tmp_score: " + err.Error())
	}

	batch := &pgx.Batch{}
	for _, s := range scores {
		queueScoreToBatch(s, batchNumber, batch)
	}
	batch.Queue(`insert into score
			(siret, siren, libelle_liste, batch, algo, periode,
			score, diff, alert, expl_selection_concerning,
			expl_selection_reassuring, macro_radar,
			redressements, alert_pre_redressements)
		select
			e.siret, t.siren, t.libelle_liste, batch, $1,
			$2, score, diff, alert, expl_selection_concerning,
			expl_selection_reassuring, macro_radar,
			redressements, alert_pre_redressements
		from tmp_score t
		inner join etablissement e on e.siren = t.siren and e.siege`, algo, now)

	batch.Queue(`insert into liste (libelle, batch, algo) values ($1, $2, $3)`, toLibelle(batchNumber), batchNumber, algo)

	results := tx.SendBatch(context.Background(), batch)
	err = results.Close()

	if err != nil {
		return errors.New("execute batch: " + err.Error())
	}
	err = tx.Commit(context.Background())
	if err != nil {
		return errors.New("commit: " + err.Error())
	}
	return nil
}

func queueScoreToBatch(s scoreFile, batchNumber string, batch *pgx.Batch) {
	sqlScore := `insert into tmp_score (siren, libelle_liste, batch, algo, score, diff, alert,
 		expl_selection_concerning, expl_selection_reassuring, macro_radar, alert_pre_redressements, redressements)
   	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	if s.ExplSelection.SelectConcerning == nil {
		s.ExplSelection.SelectConcerning = make([][]string, 0)
	}
	if s.ExplSelection.SelectReassuring == nil {
		s.ExplSelection.SelectReassuring = make([][]string, 0)
	}

	if s.MacroRadar == nil {
		s.MacroRadar = make(map[string]float64)
	}
	if s.Redressements == nil {
		s.Redressements = make([]string, 0)
	}

	batch.Queue(sqlScore,
		s.Siren,
		toLibelle(batchNumber),
		batchNumber,
		s.Algo,
		s.Score,
		s.Diff,
		s.Alert,
		s.ExplSelection.SelectConcerning,
		s.ExplSelection.SelectReassuring,
		s.MacroRadar,
		s.AlertPreRedressements,
		s.Redressements,
	)
}
