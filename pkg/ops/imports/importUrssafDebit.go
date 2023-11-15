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

type csvDebit struct {
	siret                        string
	numeroCompte                 string
	numeroEcartNegatif           string
	dateTraitement               time.Time
	partOuvriere                 float64
	partPatronale                float64
	numeroHistoriqueEcartNegatif string
	etatCompte                   int
	codeProcedureCollective      int
	periode                      [2]time.Time
	codeOperationEcartNegatif    int
	codeMotifEcartNegatif        int
	recours                      bool
}

func (d csvDebit) Values() []interface{} {
	return []interface{}{
		d.siret,
		d.numeroCompte,
		d.numeroEcartNegatif,
		d.dateTraitement,
		d.partOuvriere,
		d.partPatronale,
		d.numeroHistoriqueEcartNegatif,
		d.etatCompte,
		d.codeProcedureCollective,
		d.periode[0],
		d.periode[1],
		d.codeOperationEcartNegatif,
		d.codeMotifEcartNegatif,
		d.recours,
	}
}

func importDebit(ctx context.Context, reader *tar.Reader) (int64, error) {
	log.Print("import des débits")
	err := ensureTmpDebit(ctx)
	if err != nil {
		return 0, err
	}

	conn := db.Get()
	columns := []string{
		"siret", "numéro_compte", "numéro_écart_négatif", "date_traitement", "part_ouvrière", "part_patronale", "numéro_historique_écart_négatif", "état_compte", "code_procédure_collective", "période_start", "période_end", "code_opération_écart_négatif", "code_motif_écart_négatif", "recours",
	}
	src := copyFromDebit{
		reader: reader,
	}
	src, err = src.Init()
	if err != nil {
		return 0, err
	}
	return conn.CopyFrom(ctx, pgx.Identifier{"tmp_debit"}, columns, src)
}

type copyFromDebit struct {
	reader    *tar.Reader
	csvReader *csv.Reader
	current   *csvDebit
}

func (c copyFromDebit) Init() (copyFromDebit, error) {
	c.csvReader = csv.NewReader(c.reader)
	c.current = &csvDebit{}

	// discard header
	_, err := c.csvReader.Read()
	return c, err
}

func (c copyFromDebit) Next() bool {
	line, err := c.csvReader.Read()
	if err == io.EOF {
		return false
	}

	if err != nil {
		return false
	}

	*c.current, err = parseDebit(line)
	return true
}

func (c copyFromDebit) Values() ([]interface{}, error) {
	return c.current.Values(), c.Err()
}

func (c copyFromDebit) Err() error {
	return nil
}

func parseDebit(line []string) (csvDebit, error) {
	var debit csvDebit
	var err error

	debit.siret = line[0]
	debit.numeroCompte = line[1]
	debit.numeroEcartNegatif = line[2]
	debit.dateTraitement, err = time.Parse("2006-01-02", line[3])
	if err != nil {
		return csvDebit{}, err
	}
	debit.partOuvriere, err = strconv.ParseFloat(line[4], 64)
	if err != nil {
		return csvDebit{}, err
	}
	debit.partPatronale, err = strconv.ParseFloat(line[5], 64)
	if err != nil {
		return csvDebit{}, err
	}
	debit.numeroHistoriqueEcartNegatif = line[6]
	debit.etatCompte, err = strconv.Atoi(line[7])
	if err != nil {
		return csvDebit{}, err
	}
	debit.codeProcedureCollective, err = strconv.Atoi(line[8])
	debit.periode, err = parsePeriode(line[9])
	if err != nil {
		return csvDebit{}, err
	}

	debit.codeOperationEcartNegatif, err = strconv.Atoi(line[10])
	if err != nil {
		return csvDebit{}, err
	}
	debit.codeMotifEcartNegatif, err = strconv.Atoi(line[11])
	if err != nil {
		return csvDebit{}, err
	}
	debit.recours, err = strconv.ParseBool(line[12])
	if err != nil {
		return csvDebit{}, err
	}

	return debit, nil
}

//go:embed sql/tmpDebit.sql
var sqlTmpDebit string

func ensureTmpDebit(ctx context.Context) error {
	conn := db.Get()
	_, err := conn.Exec(ctx, sqlTmpDebit)
	return err
}
