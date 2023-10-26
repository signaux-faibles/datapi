package stats

import (
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
)

var headerStyle = excelize.Style{
	Border: []excelize.Border{{Type: "bottom",
		Color: "000000",
		Style: 2}},
	Fill: excelize.Fill{Type: "gradient", Color: []string{"FFFFFF", "4E71BE"}, Shading: 10},
	Font: nil,
	Alignment: &excelize.Alignment{
		Horizontal:  "center",
		ShrinkToFit: true,
		Vertical:    "justify",
		WrapText:    false,
	},
	Protection:    nil,
	NumFmt:        0,
	DecimalPlaces: nil,
	CustomNumFmt:  nil,
	NegRed:        false,
}

func newExcel() *excelize.File {
	return excelize.NewFile()
}

func addSheet(f *excelize.File, sheetName string) (int, error) {
	defaultSheetName := "Sheet1"
	if len(f.GetSheetList()) == 1 && f.GetSheetName(0) == defaultSheetName {
		err := f.SetSheetName(defaultSheetName, sheetName)
		if err != nil {
			return -1, errors.Wrap(err, "erreur lors du changement de nom de la première feuille")
		}
		return f.GetSheetIndex(sheetName)
	}
	return f.NewSheet(sheetName)
}

type sheetConfig[A any] interface {
	label() string
	headers() []any
	sizes() []float64
	toRow(a A) []any
	startAt() int
}

type anySheetConfig[A any] struct {
	item      A
	sheetName string
	startRow  int
	asRow     func(a A) []any
}

func (c anySheetConfig[A]) label() string {
	return c.sheetName
}

func (c anySheetConfig[A]) headers() []any {
	structure := reflect.TypeOf(c.item)
	headers := []any{}
	for i := 0; i < structure.NumField(); i++ {
		col := structure.Field(i).Tag.Get("col")
		if col != "" {
			headers = append(headers, col)
		}
	}
	return headers
}

func (c anySheetConfig[A]) sizes() []float64 {
	structure := reflect.TypeOf(c.item)
	sizes := []float64{}
	for i := 0; i < structure.NumField(); i++ {
		size := structure.Field(i).Tag.Get("size")
		if size != "" {
			sizeF, _ := strconv.ParseFloat(size, 64)
			sizes = append(sizes, float64(sizeF))
		}
	}
	return sizes
}

func (c anySheetConfig[A]) toRow(a A) []any {
	return c.asRow(a)
}

func (c anySheetConfig[A]) startAt() int {
	return c.startRow
}

func writeOneSheetToExcel[A any](
	xls *excelize.File,
	config sheetConfig[A],
	itemsToWrite chan row[A],
) error {
	sheetName := config.label()
	_, err := addSheet(xls, sheetName)
	if err != nil {
		return fmt.Errorf("erreur lors de l'ajout de la page `%s`: %w", sheetName, err)
	}

	// ajoute les styles
	headerStyleID, err := addStyle(xls, headerStyle)
	if err != nil {
		return err
	}

	// prepare writing
	writer, err := xls.NewStreamWriter(sheetName)
	if err != nil {
		return fmt.Errorf("erreur pendant la création du stream writer pour la page `%s`: %w", sheetName, err)
	}
	var flushError error
	defer func() {
		if flushError = writer.Flush(); flushError != nil {
			slog.Error("erreur à la purge du writer", slog.Any("error", err), slog.String("page", sheetName))
		}
	}()

	// configure la largeur des colonnes
	err = configureColumns(writer, config.sizes())
	if err != nil {
		return fmt.Errorf("erreur pendant la configuration des colonnes pour la page `%s`: %w", sheetName, err)
	}

	// écrit les headers
	err = writeHeaders(writer, config.headers(), headerStyleID)
	if err != nil {
		return err
	}

	// write rows
	err = writeRows(writer, config, itemsToWrite)
	if err != nil {
		return fmt.Errorf("erreur pendant l'écriture des lignes pour la page `%s`: %w", sheetName, err)
	}
	return nil
}

func addStyle(xls *excelize.File, style excelize.Style) (int, error) {
	styleID, err := xls.NewStyle(&style)
	if err != nil {
		return styleID, errors.Wrap(err, "erreur pendant l'ajout d'un style")
	}
	return styleID, nil
}

func writeHeaders(writer *excelize.StreamWriter, headers []any, style int) error {
	headersRow := []interface{}{}
	for _, header := range headers {
		cell := excelize.Cell{
			StyleID: style,
			Value:   header,
		}
		headersRow = append(headersRow, cell)
	}
	err := writer.SetRow("A1", headersRow, excelize.RowOpts{Height: 45, Hidden: false})
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'ajout de la ligne des headers")
	}
	return nil
}

func writeRows[A any](writer *excelize.StreamWriter, config sheetConfig[A], itemsToWrite chan row[A]) error {
	// write rows
	var rowIdx = config.startAt()
	if itemsToWrite != nil {
		for ligne := range itemsToWrite {
			if ligne.err != nil {
				return errors.Wrap(ligne.err, fmt.Sprint("erreur dans la ligne", rowIdx))
			}
			err := writeRow(writer, config.toRow(ligne.value), rowIdx)
			if err != nil {
				return err
			}
			rowIdx++
		}
	}
	return nil
}

func configureColumns(writer *excelize.StreamWriter, widths []float64) error {
	for i, width := range widths {
		col := i + 1
		err := writer.SetColWidth(col, col, width)
		if err != nil {
			return fmt.Errorf("erreur pendant la configuration de la largeur de la colonne '%d' : %w", col, err)
		}
	}
	return nil
}

func writeRow(f *excelize.StreamWriter, values []any, row int) error {
	cell, err := excelize.CoordinatesToCellName(1, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant la récupération des coordonnées")
	}
	err = f.SetRow(cell, values)
	if err != nil {
		return fmt.Errorf("erreur pendant l'écriture de la ligne : %w", err)
	}
	return nil
}

func exportTo(output io.Writer, xlsx *excelize.File) error {
	err := xlsx.Write(output, excelize.Options{RawCellValue: true})
	if err != nil {
		slog.Error("erreur pendant l'écriture du fichier xlsx")
		return err
	}
	return nil
}
