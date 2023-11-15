package imports

import (
	"archive/tar"
	"context"
	"datapi/pkg/db"
	_ "embed"
	"encoding/csv"
	"fmt"
	"github.com/jackc/pgx/v5"
	"io"
	"log"
	"strconv"
	"time"
)

type csvEffectif struct {
	siret        string
	numéroCompte string
	période      time.Time
	effectif     int
	err          error
}

func (d csvEffectif) Values() []interface{} {
	return []interface{}{
		d.siret,
		d.numéroCompte,
		d.période,
		d.effectif,
	}
}

func importEffectif(ctx context.Context, reader *tar.Reader) (int64, error) {
	log.Print("import des effectifs")
	err := ensureTmpEffectif(ctx)
	if err != nil {
		return 0, err
	}

	conn := db.Get()
	columns := []string{
		"siret", "numéro_compte", "période", "effectif",
	}
	src := copyFromEffectif{
		reader: reader,
	}
	src, err = src.Init()
	if err != nil {
		return 0, err
	}
	return conn.CopyFrom(ctx, pgx.Identifier{"tmp_effectif"}, columns, src)
}

type copyFromEffectif struct {
	reader    *tar.Reader
	csvReader *csv.Reader
	current   *csvEffectif
}

func (c copyFromEffectif) Init() (copyFromEffectif, error) {
	c.csvReader = csv.NewReader(c.reader)
	c.current = &csvEffectif{}

	// discard header
	_, err := c.csvReader.Read()
	return c, err
}

func (c copyFromEffectif) Next() bool {
	line, err := c.csvReader.Read()
	if err == io.EOF {
		return false
	}

	if err != nil {
		return false
	}

	*c.current = parseEffectif(line)
	if c.current.err != nil {
		fmt.Println(c.current.err)
	}
	return true
}

func (c copyFromEffectif) Values() ([]interface{}, error) {
	return c.current.Values(), c.Err()
}

func (c copyFromEffectif) Err() error {
	return c.current.err
}

func parseEffectif(line []string) csvEffectif {
	var effectif csvEffectif
	var err error

	effectif.siret = line[0]
	effectif.numéroCompte = line[1]
	effectif.période, err = time.Parse("2006-01-02", line[2])
	if err != nil {
		return csvEffectif{err: err}
	}
	effectif.effectif, err = strconv.Atoi(line[3])
	if err != nil {
		return csvEffectif{err: err}
	}
	return effectif
}

//go:embed sql/tmpEffectif.sql
var sqlTmpEffectif string

func ensureTmpEffectif(ctx context.Context) error {
	conn := db.Get()
	_, err := conn.Exec(ctx, sqlTmpEffectif)
	return err
}
