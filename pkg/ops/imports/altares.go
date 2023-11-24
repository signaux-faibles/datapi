package imports

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"

	"datapi/pkg/core"
	"datapi/pkg/db"
)

type Paydex struct {
	Siren        core.Siren
	Paydex       *string
	JoursRetard  *int
	Fournisseurs *int
	Encours      *float64
	FPI30        *int
	FPI90        *int
	DateValeur   *time.Time
	err          error
}

func (p Paydex) tuple() []interface{} {
	return []interface{}{
		p.Siren, p.Paydex, p.JoursRetard, p.Fournisseurs, p.Encours, p.FPI30, p.FPI90, p.DateValeur,
	}
}

func PaydexReader(ctx context.Context, filename string) (chan Paydex, error) {
	output := make(chan Paydex)
	zipReader, err := zip.OpenReader(filename)
	if err != nil {
		close(output)
		return output, err
	}

	go func() {
		defer close(output)
		for _, zf := range zipReader.File {
			file, errOpenZip := zf.Open()
			if errOpenZip != nil {
				output <- Paydex{err: errOpenZip}
				continue
			}
			reader := csv.NewReader(file)
			reader.Comma = ','
			headers, errReadHeaders := reader.Read()
			if errReadHeaders != nil {
				output <- Paydex{err: errReadHeaders}
				continue
			}
			if !acceptPaydexHeaders(headers) {
				continue
			}
			for {
				line, errCsvReader := reader.Read()
				if errCsvReader == io.EOF {
					break
				} else if errCsvReader != nil {
					output <- Paydex{err: errCsvReader}
					break
				}
				select {
				case <-ctx.Done():
					return
				case output <- parsePaydex(line):
				}
			}
		}
	}()

	return output, nil
}

func parsePaydex(line []string) Paydex {
	if len(line) != 10 {
		return Paydex{
			err: csv.ErrFieldCount,
		}
	}

	var paydex Paydex

	if core.Siren(line[0]).IsValid() {
		paydex.Siren = core.Siren(line[0])
	} else {
		return Paydex{
			err: fmt.Errorf("%s is not a valid Siren", line[0]),
		}
	}

	dateValeur, err := time.Parse("2006-01-02", line[9])
	if err != nil {
		return Paydex{
			err: fmt.Errorf("%s is not a valid DATE_VALEUR value: %s", line[2], err.Error()),
		}
	}
	paydex.DateValeur = &dateValeur

	paydex.Paydex = parsePstring(line[2])

	paydex.JoursRetard, err = parsePint(line[4])
	if err != nil {
		return Paydex{
			err: fmt.Errorf("%s is not a valid NBR_JRS_RETARD value: %s", line[2], err.Error()),
		}
	}

	paydex.Fournisseurs, err = parsePint(line[5])
	if err != nil {
		return Paydex{
			err: fmt.Errorf("%s is not a valid NBR_FOURNISSEURS value: %s", line[2], err.Error()),
		}
	}

	paydex.Encours, err = parsePfloat(line[6])
	if err != nil {
		return Paydex{
			err: fmt.Errorf("%s is not a valid ENCOURS_ETUDIES value: %s", line[2], err.Error()),
		}
	}

	paydex.FPI30, err = parsePint(line[7])
	if err != nil {
		return Paydex{
			err: fmt.Errorf("%s is not a valid NBR_EXPERIENCES_PAIEMENT value: %s", line[2], err.Error()),
		}
	}

	paydex.FPI90, err = parsePint(line[8])
	if err != nil {
		return Paydex{
			err: fmt.Errorf("%s is not a valid NBR_EXPERIENCES_PAIEMENT value: %s", line[2], err.Error()),
		}
	}

	return paydex
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

func acceptPaydexHeaders(fields []string) bool {
	bomUTF8 := string([]byte{0xEF, 0xBB, 0xBF})
	headers := []string{bomUTF8 + "siren", "état_organisation", "code_paydex", "paydex", "nbr_jrs_retard", "nbr_fournisseurs", "encours_étudiés", "note_100_alerteur_plus_30", "note_100_alerteur_plus_90_jours", "date_valeur"}
	if len(headers) != len(fields) {
		slog.Info("Fichier rejeté, header non conforme", slog.Int("expected_length", len(headers)), slog.Int("actual_length", len(fields)))
		return false
	}
	for index := range fields {
		if fields[index] != headers[index] {
			slog.Info("Fichier rejeté, header non conforme", slog.Int("index", index), slog.String("expected", headers[index]), slog.String("actual", fields[index]))
			return false
		}
	}
	return true
}

func importPaydex(ctx context.Context) error {
	paydexFilePath := viper.GetString("source.paydexpath")
	paydexReader, err := PaydexReader(ctx, paydexFilePath)
	if err != nil {
		return err
	}
	copyFromPaydex := CopyFromPaydex{
		PaydexParser: paydexReader,
		Current:      &Paydex{},
		Count:        new(int),
	}
	err = dropEntreprisePaydexIndex(ctx)
	if err != nil {
		return err
	}
	err = truncatePaydexTable(ctx)
	if err != nil {
		return err
	}
	err = copyPaydex(ctx, copyFromPaydex)
	if err != nil {
		return err
	}
	return createEntreprisePaydexIndex(ctx)
}

func truncatePaydexTable(ctx context.Context) error {
	conn := db.Get()
	slog.Info("truncating table")
	sql := `truncate table entreprise_paydex;`
	_, err := conn.Exec(ctx, sql)
	return err
}
func dropEntreprisePaydexIndex(ctx context.Context) error {
	conn := db.Get()
	slog.Info("dropping database index")
	sql := `drop index if exists idx_entreprise_paydex_siren;`
	_, err := conn.Exec(ctx, sql)
	return err
}

func createEntreprisePaydexIndex(ctx context.Context) error {
	conn := db.Get()
	slog.Info("creating database index")
	sql := `create index idx_entreprise_paydex_siren on entreprise_paydex (siren, date_valeur);`
	_, err := conn.Exec(ctx, sql)
	return err
}

func copyPaydex(ctx context.Context, copyFromSource pgx.CopyFromSource) error {
	conn := db.Get()
	headers := []string{
		"siren",
		"paydex",
		"jours_retard",
		"fournisseurs",
		"en_cours",
		"fpi_30",
		"fpi_90",
		"date_valeur",
	}
	identifier := pgx.Identifier{"entreprise_paydex"}

	_, err := conn.CopyFrom(ctx, identifier, headers, copyFromSource)
	return err
}

type CopyFromPaydex struct {
	PaydexParser chan Paydex
	Current      *Paydex
	Count        *int
}

func (c CopyFromPaydex) Increment() {
	*c.Count = *c.Count + 1
}

func (c CopyFromPaydex) Next() bool {
	var ok bool
	select {
	case *c.Current, ok = <-c.PaydexParser:
		if ok {
			if c.Current.err != nil {
				return c.Next()
			}
			c.Increment()
			if *c.Count%100000 == 0 {
				slog.Info("paydex objects copied", slog.Int("counter", *c.Count))
			}
		}
		return ok
	}
}

func (c CopyFromPaydex) Err() error { return nil }
func (c CopyFromPaydex) Values() ([]interface{}, error) {
	return (*c.Current).tuple(), nil
}
