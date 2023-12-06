package stats

import (
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_write_statsAccessLogsSheet(t *testing.T) {
	xls := newExcel()
	first, items := createFakeActivitesChan(3, createFakeAccessLog)

	// écriture dans le fichier

	sheetConf := accessLogSheetConfig()
	err := writeOneSheetToExcel(xls, sheetConf, items)
	require.NoError(t, err)

	// sauvegarde dans un fichier
	output, err := os.CreateTemp(os.TempDir(), t.Name()+"_*.xlsx")
	require.NoError(t, err)
	slog.Info("export du access log", slog.String("filename", output.Name()))
	err = xls.Write(output)
	require.NoError(t, err)

	// vérifications
	rows, err := xls.Rows(sheetConf.label())
	require.NoError(t, err)

	require.True(t, rows.Next())
	headerValues, err := rows.Columns()
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"date", "chemin", "méthode", "utilisateur", "segment", "rôles"}, headerValues)

	require.True(t, rows.Next())
	firstRowDataValues, err := rows.Columns()
	require.NoError(t, err)

	//
	expected := sheetConf.toRow(first)
	expected[0] = first.date.Format(time.DateTime)
	assert.ElementsMatch(t, expected, firstRowDataValues)
}

func createFakeAccessLog() accessLog {
	return accessLog{
		date:     fakeTU.Time().TimeBetween(time.Now().AddDate(-3, 0, 0), time.Now()),
		path:     fakeTU.Directory().Directory(fakeTU.IntBetween(1, 3)),
		method:   fakeTU.Internet().HTTPMethod(),
		username: fakeTU.Internet().User(),
		segment:  fakeTU.RandomStringElement([]string{"bdf", "dgfip", "sf"}),
		roles:    fakeTU.Lorem().Words(fakeTU.IntBetween(1, 99)),
	}
}

func Test_accessLogSheetConfig(t *testing.T) {
	structure := reflect.TypeOf(accessLog{})
	headers := []string{}
	sizes := []float32{}
	for i := 0; i < structure.NumField(); i++ {
		col := structure.Field(i).Tag.Get("col")
		if col != "" {
			headers = append(headers, col)
		}
		size := structure.Field(i).Tag.Get("size")
		if size != "" {
			sizeF, _ := strconv.ParseFloat(size, 32)
			sizes = append(sizes, float32(sizeF))
		}
	}
	slog.Info("headers", slog.Any("headers", headers))
	slog.Info("sizes", slog.Any("sizes", sizes))
}

func Test_accessLogsSelector_sql(t *testing.T) {
	type fields struct {
		from time.Time
		to   time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := accessLogsSelector{
				from: tt.fields.from,
				to:   tt.fields.to,
			}
			assert.Equalf(t, tt.want, a.sql(), "sql()")
		})
	}
}

func Test_accessLogsSelector_sqlArgs(t *testing.T) {
	type fields struct {
		from time.Time
		to   time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   []any
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := accessLogsSelector{
				from: tt.fields.from,
				to:   tt.fields.to,
			}
			assert.Equalf(t, tt.want, a.sqlArgs(), "sqlArgs()")
		})
	}
}

func Test_accessLogsSelector_toItem(t *testing.T) {
	type fields struct {
		from time.Time
		to   time.Time
	}
	type args struct {
		rows pgx.Rows
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    accessLog
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := accessLogsSelector{
				from: tt.fields.from,
				to:   tt.fields.to,
			}
			got, err := a.toItem(tt.args.rows)
			if !tt.wantErr(t, err, fmt.Sprintf("toItem(%v)", tt.args.rows)) {
				return
			}
			assert.Equalf(t, tt.want, got, "toItem(%v)", tt.args.rows)
		})
	}
}

func Test_newAccessLogsSelector(t *testing.T) {
	type args struct {
		from time.Time
		to   time.Time
	}
	tests := []struct {
		name string
		args args
		want accessLogsSelector
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, newAccessLogsSelector(tt.args.from, tt.args.to), "newAccessLogsSelector(%v, %v)", tt.args.from, tt.args.to)
		})
	}
}

func Test_toRow(t *testing.T) {
	type args struct {
		ligne accessLog
	}
	tests := []struct {
		name string
		args args
		want []any
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, toRow(tt.args.ligne), "toRow(%v)", tt.args.ligne)
		})
	}
}
