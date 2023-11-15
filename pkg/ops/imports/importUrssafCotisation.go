package imports

import (
	"archive/tar"
	"context"
	"datapi/pkg/db"
	_ "embed"
	"encoding/csv"
	"github.com/jackc/pgx/v5"
	"io"
	"log"
	"strconv"
	"time"
)

type csvCotisation struct {
	siret        string
	numeroCompte string
	période      [2]time.Time
	encaissé     float64
	du           float64
	err          error
}

func (d csvCotisation) Values() []interface{} {
	return []interface{}{
		d.siret,
		d.numeroCompte,
		d.période[0],
		d.période[1],
		d.encaissé,
		d.du,
	}
}

func importCotisation(ctx context.Context, reader *tar.Reader) (int64, error) {
	log.Print("import des cotisation")
	err := ensureTmpCotisation(ctx)
	if err != nil {
		return 0, err
	}

	conn := db.Get()
	columns := []string{
		"siret", "numéro_compte", "période_start", "période_end", "encaissé", "du",
	}
	src := copyFromCotisation{
		reader: reader,
	}
	src, err = src.Init()
	if err != nil {
		return 0, err
	}
	return conn.CopyFrom(ctx, pgx.Identifier{"tmp_cotisation"}, columns, src)
}

type copyFromCotisation struct {
	reader    *tar.Reader
	csvReader *csv.Reader
	current   *csvCotisation
}

func (c copyFromCotisation) Init() (copyFromCotisation, error) {
	c.csvReader = csv.NewReader(c.reader)
	c.current = &csvCotisation{}

	// discard header
	_, err := c.csvReader.Read()
	return c, err
}

func (c copyFromCotisation) Next() bool {
	line, err := c.csvReader.Read()
	if err == io.EOF {
		return false
	}

	if err != nil {
		return false
	}

	*c.current = parseCotisation(line)
	return c.current.err == nil
}

func (c copyFromCotisation) Values() ([]interface{}, error) {
	return c.current.Values(), c.current.err
}

func (c copyFromCotisation) Err() error {
	return c.current.err
}

func parseCotisation(line []string) csvCotisation {
	var cotisation csvCotisation
	var err error

	cotisation.siret = line[0]
	cotisation.numeroCompte = line[1]
	cotisation.période, err = parsePeriode(line[2])
	if err != nil {
		return csvCotisation{err: err}
	}
	cotisation.encaissé, err = strconv.ParseFloat(line[3], 64)
	if err != nil {
		return csvCotisation{err: err}
	}
	cotisation.du, err = strconv.ParseFloat(line[4], 64)
	if err != nil {
		return csvCotisation{err: err}
	}
	return cotisation
}

//go:embed sql/tmpCotisation.sql
var sqlTmpCotisation string

func ensureTmpCotisation(ctx context.Context) error {
	conn := db.Get()
	_, err := conn.Exec(ctx, sqlTmpCotisation)
	return err
}
