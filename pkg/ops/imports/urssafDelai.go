package imports

import (
	"archive/tar"
	"context"
	_ "embed"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"

	"datapi/pkg/db"
)

type csvDelai struct {
	siret             string
	numéroCompte      string
	numéroContentieux string
	dateCréation      time.Time
	dateÉchéance      time.Time
	duréeDélai        int
	dénomination      string
	indic6mois        bool
	annéeCréation     int
	montantÉchéancier float64
	stade             string
	action            string
}

func (d csvDelai) Values() []interface{} {
	return []interface{}{
		d.siret,
		d.numéroCompte,
		d.numéroContentieux,
		d.dateCréation,
		d.dateÉchéance,
		d.duréeDélai,
		d.dénomination,
		d.indic6mois,
		d.annéeCréation,
		d.montantÉchéancier,
		d.stade,
		d.action,
	}
}

func importDelai(ctx context.Context, reader *tar.Reader) (int64, error) {
	slog.Info("import des délais")
	err := ensureTmpDelai(ctx)
	if err != nil {
		return 0, err
	}

	conn := db.Get()
	columns := []string{
		"siret", "numéro_compte", "numéro_contentieux", "date_création", "date_échéance", "durée_délai", "dénomination", "indic_6mois", "année_création", "montant_échéancier", "stade", "action",
	}
	src := copyFromDelai{
		reader: reader,
	}
	src, err = src.Init()
	if err != nil {
		return 0, err
	}
	return conn.CopyFrom(ctx, pgx.Identifier{"tmp_delai"}, columns, src)
}

type copyFromDelai struct {
	reader    *tar.Reader
	csvReader *csv.Reader
	current   *csvDelai
}

func (c copyFromDelai) Init() (copyFromDelai, error) {
	c.csvReader = csv.NewReader(c.reader)
	c.current = &csvDelai{}

	// discard header
	_, err := c.csvReader.Read()
	return c, err
}

func (c copyFromDelai) Next() bool {
	line, err := c.csvReader.Read()
	if err == io.EOF {
		return false
	}

	if err != nil {
		return false
	}

	*c.current, err = parseDelai(line)
	if err != nil {
		fmt.Println(err)
	}
	return true
}

func (c copyFromDelai) Values() ([]interface{}, error) {
	return c.current.Values(), c.Err()
}

func (c copyFromDelai) Err() error {
	return nil
}

func parseDelai(line []string) (csvDelai, error) {
	var delai csvDelai
	var err error

	delai.siret = line[0]
	delai.numéroCompte = line[1]
	delai.numéroContentieux = line[2]
	delai.dateCréation, err = time.Parse("2006-01-02", line[3])
	if err != nil {
		return csvDelai{}, err
	}
	delai.dateÉchéance, err = time.Parse("2006-01-02", line[4])
	if err != nil {
		return csvDelai{}, err
	}
	delai.duréeDélai, err = strconv.Atoi(line[5])
	if err != nil {
		return csvDelai{}, err
	}
	delai.dénomination = line[6]
	delai.indic6mois = (line[7] == "SUP")

	delai.annéeCréation, err = strconv.Atoi(line[8])
	if err != nil {
		return csvDelai{}, err
	}

	delai.montantÉchéancier, err = strconv.ParseFloat(line[9], 64)
	if err != nil {
		return csvDelai{}, err
	}

	delai.stade = line[10]
	delai.action = line[11]

	return delai, nil
}

//go:embed sql/tmpDelai.sql
var sqlTmpDelai string

func ensureTmpDelai(ctx context.Context) error {
	conn := db.Get()
	_, err := conn.Exec(ctx, sqlTmpDelai)
	return err
}
