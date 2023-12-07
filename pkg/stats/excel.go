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

var styleHeader = excelize.Style{
	Border: []excelize.Border{{Type: "bottom",
		Color: "000000",
		Style: 2}},
	Fill: excelize.Fill{Type: "pattern", Color: []string{"4E71BE"}, Pattern: 1},
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
	headers() ([]any, error)
	sizes() []float64
	toRow(a A) []any
	formatters() map[int]excelize.Style
}

type anySheetConfig[A any] struct {
	item      A
	sheetName string
	asRow     func(a A) []any
}

func (c anySheetConfig[A]) label() string {
	return c.sheetName
}

func (c anySheetConfig[A]) headers() ([]any, error) {
	structure := reflect.TypeOf(c.item)
	headers := []any{}
	for i := 0; i < structure.NumField(); i++ {
		col := structure.Field(i).Tag.Get("col")
		if col == "" {
			return nil, errors.New("erreur pendant la génération des headers : le tag `col` n'est pas défini")
		}
		headers = append(headers, col)
	}
	return headers, nil
}

func (c anySheetConfig[A]) sizes() []float64 {
	structure := reflect.TypeOf(c.item)
	sizes := []float64{}
	for i := 0; i < structure.NumField(); i++ {
		size := float64(24)
		configuredSize := structure.Field(i).Tag.Get("size")
		if configuredSize != "" {
			sizeF, _ := strconv.ParseFloat(configuredSize, 64)
			size = sizeF
		}
		sizes = append(sizes, size)
	}
	return sizes
}

func (c anySheetConfig[A]) toRow(a A) []any {
	return c.asRow(a)
}

func (c anySheetConfig[A]) formatters() map[int]excelize.Style {
	structure := reflect.TypeOf(c.item)
	formatters := make(map[int]excelize.Style)
	for i := 0; i < structure.NumField(); i++ {
		dateFormat := structure.Field(i).Tag.Get("dateFormat")
		if dateFormat != "" {
			formatters[i] = excelize.Style{CustomNumFmt: &dateFormat}
		}
	}
	return formatters
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

	// ajoute les formatters
	headerStyleID, err := addStyle(xls, styleHeader)
	if err != nil {
		return err
	}
	xlsStyleIDS, err := addStyles(xls, config.formatters())
	if err != nil {
		return err
	}

	// ajoute les auto filters
	err = setAutoFilter(xls, sheetName, config)
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
	headers, err := config.headers()
	err = writeHeaders(writer, headers, headerStyleID)
	if err != nil {
		return err
	}

	// write rows
	err = writeRows(writer, config, itemsToWrite, xlsStyleIDS)
	if err != nil {
		return fmt.Errorf("erreur pendant l'écriture des lignes pour la page `%s`: %w", sheetName, err)
	}
	return flushError
}

func setAutoFilter[A any](xls *excelize.File, sheetName string, cf sheetConfig[A]) error {
	var ops []excelize.AutoFilterOptions
	headers, err := cf.headers()
	if err != nil {
		return errors.Wrap(err, "erreur pendant la configuration des auto filters")
	}
	name, err := excelize.CoordinatesToCellName(len(headers), 1)
	if err != nil {
		return errors.Wrap(err, "erreur pendant la configuration des auto filters (erreur de récupération de coordonnées)")
	}
	err = xls.AutoFilter(sheetName, "A1:"+name, ops)
	if err != nil {
		return errors.Wrap(err, "erreur pendant la configuration des auto filters")
	}
	return nil
}

func addStyles(xls *excelize.File, stylesMap map[int]excelize.Style) (map[int]int, error) {
	xlsStyleIDS := make(map[int]int)
	for idx, style := range stylesMap {
		styleID, err := addStyle(xls, style)
		if err != nil {
			return nil, err
		}
		xlsStyleIDS[idx] = styleID
	}
	return xlsStyleIDS, nil
}

func addStyle(xls *excelize.File, style excelize.Style) (int, error) {
	styleID, err := xls.NewStyle(&style)
	if err != nil {
		return styleID, errors.Wrap(err, "erreur pendant l'ajout d'un style")
	}
	return styleID, nil
}

func writeHeaders(writer *excelize.StreamWriter, headers []any, style int) error {
	headersRow := []any{}
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

func writeRows[A any](
	writer *excelize.StreamWriter,
	config sheetConfig[A],
	itemsToWrite chan row[A],
	styles map[int]int,
) error {
	// write rows
	var rowIdx = 2 // le header est déjà écrit
	if itemsToWrite != nil {
		for ligne := range itemsToWrite {
			if ligne.err != nil {
				return errors.Wrap(ligne.err, fmt.Sprint("erreur dans la ligne", rowIdx))
			}
			err := writeRow(writer, config.toRow(ligne.value), rowIdx, styles)
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

func writeRow(f *excelize.StreamWriter, values []any, row int, styles map[int]int) error {
	cell, err := excelize.CoordinatesToCellName(1, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant la récupération des coordonnées")
	}
	cells := []any{}
	for idx, v := range values {
		cellValue := v
		styleID, ok := styles[idx]
		if ok {
			cellValue = excelize.Cell{
				StyleID: styleID,
				Value:   v,
			}
		}
		cells = append(cells, cellValue)
	}
	err = f.SetRow(cell, cells)
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
