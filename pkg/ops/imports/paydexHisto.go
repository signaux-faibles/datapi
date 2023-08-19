package imports

import (
	"archive/zip"
	"context"
	"datapi/pkg/core"
	"datapi/pkg/db"
	"encoding/csv"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
	"io"
	"log"
	"strconv"
	"time"
)

type PaydexHisto struct {
	Siren               core.Siren
	Paydex              *string
	JoursRetard         *int
	Fournisseurs        *int
	Encours             *float64
	ExperiencesPaiement *int
	FPI30               *int
	FPI90               *int
	DateValeur          *time.Time
	err                 error
}

func (p PaydexHisto) tuple() []interface{} {
	return []interface{}{
		p.Siren, p.Paydex, p.JoursRetard, p.Fournisseurs, p.Encours, p.ExperiencesPaiement, p.FPI30, p.FPI90, p.DateValeur,
	}
}

func PaydexHistoReader(ctx context.Context, filename string) (chan PaydexHisto, error) {
	zipReader, err := zip.OpenReader(filename)
	output := make(chan PaydexHisto)
	if err != nil {
		close(output)
		return output, err
	}

	go func() {
		defer close(output)
		for _, zf := range zipReader.File {
			file, errOpenZip := zf.Open()
			if errOpenZip != nil {
				output <- PaydexHisto{err: errOpenZip}
				continue
			}
			reader := csv.NewReader(file)
			reader.Comma = ';'
			headers, errReadHeaders := reader.Read()
			if errReadHeaders != nil {
				output <- PaydexHisto{err: errReadHeaders}
				continue
			}
			if !checkHistoHeaders(headers) {
				continue
			}
			for {
				line, errCsvReader := reader.Read()
				if errCsvReader == io.EOF {
					break
				} else if errCsvReader != nil {
					output <- PaydexHisto{err: errCsvReader}
					break
				}
				select {
				case <-ctx.Done():
					return
				case output <- parsePaydexHisto(line):
				}
			}
		}
	}()

	return output, nil
}

func parsePaydexHisto(line []string) PaydexHisto {
	if len(line) != 11 {
		return PaydexHisto{
			err: csv.ErrFieldCount,
		}
	}

	var paydexHisto PaydexHisto

	if core.Siren(line[0]).IsValid() {
		paydexHisto.Siren = core.Siren(line[0])
	} else {
		return PaydexHisto{
			err: fmt.Errorf("%s is not a valid Siren", line[0]),
		}
	}

	dateValeur, err := time.Parse("02/01/2006", line[10])
	if err != nil {
		return PaydexHisto{
			err: fmt.Errorf("%s is not a valid DATE_VALEUR value: %s", line[2], err.Error()),
		}
	}
	paydexHisto.DateValeur = &dateValeur

	paydexHisto.Paydex = parsePstring(line[2])

	paydexHisto.JoursRetard, err = parsePint(line[4])
	if err != nil {
		return PaydexHisto{
			err: fmt.Errorf("%s is not a valid NBR_JRS_RETARD value: %s", line[2], err.Error()),
		}
	}

	paydexHisto.Fournisseurs, err = parsePint(line[5])
	if err != nil {
		return PaydexHisto{
			err: fmt.Errorf("%s is not a valid NBR_FOURNISSEURS value: %s", line[2], err.Error()),
		}
	}

	paydexHisto.Encours, err = parsePfloat(line[6])
	if err != nil {
		return PaydexHisto{
			err: fmt.Errorf("%s is not a valid ENCOURS_ETUDIES value: %s", line[2], err.Error()),
		}
	}

	paydexHisto.ExperiencesPaiement, err = parsePint(line[7])
	if err != nil {
		return PaydexHisto{
			err: fmt.Errorf("%s is not a valid NBR_EXPERIENCES_PAIEMENT value: %s", line[2], err.Error()),
		}
	}

	paydexHisto.FPI30, err = parsePint(line[8])
	if err != nil {
		return PaydexHisto{
			err: fmt.Errorf("%s is not a valid NBR_EXPERIENCES_PAIEMENT value: %s", line[2], err.Error()),
		}
	}

	paydexHisto.FPI90, err = parsePint(line[9])
	if err != nil {
		return PaydexHisto{
			err: fmt.Errorf("%s is not a valid NBR_EXPERIENCES_PAIEMENT value: %s", line[2], err.Error()),
		}
	}

	return paydexHisto
}

func parsePfloat(value string) (*float64, error) {
	if value == "" {
		return nil, nil
	}
	f, err := strconv.ParseFloat(value, 64)
	return &f, err
}

func parsePstring(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func parsePint(value string) (*int, error) {
	if value == "" {
		return nil, nil
	}
	i, err := strconv.Atoi(value)
	return &i, err
}

func checkHistoHeaders(fields []string) bool {
	bomUTF8 := string([]byte{0xEF, 0xBB, 0xBF})
	headers := []string{bomUTF8 + "SIREN", "ETAT_ORGANISATION", "CODE_PAYDEX", "PAYDEX", "NBR_JRS_RETARD", "NBR_FOURNISSEURS", "ENCOURS_ETUDIES", "NBR_EXPERIENCES_PAIEMENT", "NOTE100_ALERTEUR_PLUS_30", "NOTE100_ALERTEUR_PLUS_90_JOURS", "DATE_VALEUR"}
	if len(headers) != len(fields) {
		return false
	}
	for index := range fields {
		if fields[index] != headers[index] {
			return false
		}
	}
	return true
}

func importPaydexHisto(ctx context.Context) error {
	paydexHistoFilePath := viper.GetString("source.paydexHistoPath")
	paydexHistoReader, err := PaydexHistoReader(ctx, paydexHistoFilePath)
	if err != nil {
		return err
	}
	copyFromPaydexHisto := CopyFromPaydexHisto{
		PaydexHistoParser: paydexHistoReader,
		Current:           &PaydexHisto{},
		Count:             new(int),
	}
	err = dropEntreprisePaydexIndex(ctx)
	if err != nil {
		return err
	}
	err = truncatePaydexTable(ctx)
	if err != nil {
		return err
	}
	err = copyPaydexHisto(ctx, copyFromPaydexHisto)
	if err != nil {
		return err
	}
	return createEntreprisePaydexIndex(ctx)
}

func truncatePaydexTable(ctx context.Context) error {
	conn := db.Get()
	log.Println("truncating table")
	sql := `truncate table entreprise_paydex;`
	_, err := conn.Exec(ctx, sql)
	return err
}
func dropEntreprisePaydexIndex(ctx context.Context) error {
	conn := db.Get()
	log.Println("dropping database index")
	sql := `drop index if exists idx_entreprise_paydex_siren;`
	_, err := conn.Exec(ctx, sql)
	return err
}

func createEntreprisePaydexIndex(ctx context.Context) error {
	conn := db.Get()
	log.Println("creating database index")
	sql := `create index idx_entreprise_paydex_siren on entreprise_paydex (siren);`
	_, err := conn.Exec(ctx, sql)
	return err
}

func copyPaydexHisto(ctx context.Context, copyFromSource pgx.CopyFromSource) error {
	conn := db.Get()
	headers := []string{
		"siren",
		"paydex",
		"jours_retard",
		"fournisseurs",
		"en_cours",
		"experiences_paiement",
		"fpi_30",
		"fpi_90",
		"date_valeur",
	}
	identifier := pgx.Identifier{"entreprise_paydex"}

	_, err := conn.CopyFrom(ctx, identifier, headers, copyFromSource)
	return err
}

type CopyFromPaydexHisto struct {
	PaydexHistoParser chan PaydexHisto
	Current           *PaydexHisto
	Count             *int
}

func (c CopyFromPaydexHisto) Increment() {
	*c.Count = *c.Count + 1
}

func (c CopyFromPaydexHisto) Next() bool {
	var ok bool
	select {
	case *c.Current, ok = <-c.PaydexHistoParser:
		if ok {
			if c.Current.err != nil {
				return c.Next()
			}
			c.Increment()
			if *c.Count%100000 == 0 {
				log.Printf("%d paydexHisto objects copied\n", *c.Count)
			}
		}
		return ok
	}
}

func (c CopyFromPaydexHisto) Err() error { return nil }
func (c CopyFromPaydexHisto) Values() ([]interface{}, error) {
	return (*c.Current).tuple(), nil
}
