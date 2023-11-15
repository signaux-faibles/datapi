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
	"time"
)

type csvProcol struct {
	siret        string
	dateEffet    time.Time
	actionProcol string
	stadeProcol  string
	err          error
}

func (d csvProcol) Values() []interface{} {
	return []interface{}{
		d.siret,
		d.dateEffet,
		d.actionProcol,
		d.stadeProcol,
	}
}

func importProcol(ctx context.Context, reader *tar.Reader) (int64, error) {
	log.Print("import des Procol")
	err := ensureTmpProcol(ctx)
	if err != nil {
		return 0, err
	}

	conn := db.Get()
	columns := []string{
		"siret", "date_effet", "action_procol", "stade_procol",
	}
	src := copyFromProcol{
		reader: reader,
	}
	src, err = src.Init()
	if err != nil {
		return 0, err
	}
	return conn.CopyFrom(ctx, pgx.Identifier{"tmp_procol"}, columns, src)
}

type copyFromProcol struct {
	reader    *tar.Reader
	csvReader *csv.Reader
	current   *csvProcol
}

func (c copyFromProcol) Init() (copyFromProcol, error) {
	c.csvReader = csv.NewReader(c.reader)
	c.current = &csvProcol{}

	// discard header
	_, err := c.csvReader.Read()
	return c, err
}

func (c copyFromProcol) Next() bool {
	line, err := c.csvReader.Read()
	if err == io.EOF {
		return false
	}

	if err != nil {
		return false
	}

	*c.current = parseProcol(line)
	return c.current.err == nil
}

func (c copyFromProcol) Values() ([]interface{}, error) {
	return c.current.Values(), c.current.err
}

func (c copyFromProcol) Err() error {
	return c.current.err
}

func parseProcol(line []string) csvProcol {
	var procol csvProcol
	var err error

	procol.siret = line[0]
	procol.dateEffet, err = time.Parse("2006-01-02", line[1])
	if err != nil {
		return csvProcol{err: err}
	}
	procol.actionProcol = line[2]
	procol.stadeProcol = line[3]

	return procol
}

//go:embed sql/tmpProcol.sql
var sqlTmpProcol string

func ensureTmpProcol(ctx context.Context) error {
	conn := db.Get()
	_, err := conn.Exec(ctx, sqlTmpProcol)
	return err
}
