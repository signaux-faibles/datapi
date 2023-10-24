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

type rowWriter[Row any] func(f *excelize.File, sheetName string, ligne Row, row int) error

func writeOneSheetToExcel2[A any](
	xls *excelize.File,
	config sheetConfig[A],
	itemsToWrite chan row[A],
) error {
	_, err := addSheet(xls, config.label())
	if err != nil {
		return fmt.Errorf("erreur lors de l'ajout de la page `%s`: %w", config.label(), err)
	}
	err = writeHeaders(xls, config)
	if err != nil {
		return fmt.Errorf("erreur lors de l'écriture des headers : %w", err)
	}
	var rowIdx = config.startAt()
	if itemsToWrite != nil {
		for ligne := range itemsToWrite {
			if ligne.err != nil {
				return ligne.err
			}
			err := writeRow(xls, config.label(), config.toRow(ligne.value), rowIdx)
			if err != nil {
				return err
			}
			rowIdx++
		}
	}
	return nil
}

func writeHeaders[A any](xls *excelize.File, config sheetConfig[A]) error {
	headers := config.headers()
	err := writeRow(xls, config.label(), headers, 1)
	if err != nil {
		return errors.Wrap(err, "erreur pendant la récupération du nom de la cellule")
	}
	//err = xls.SetRowOutlineLevel(config.label(), 1, 1)
	//if err != nil {
	//	return errors.Wrap(err, "erreur pendant la définition du niveau d'outline")
	//}
	err = xls.SetRowHeight(config.label(), 1, 32)
	if err != nil {
		return errors.Wrap(err, "erreur pendant la définition de la hauteur de colonne")
	}
	headerStyleID, err := xls.NewStyle(&headerStyle)
	if err != nil {
		return errors.Wrap(err, "erreur pendant la création du style de header")
	}
	err = xls.SetRowStyle(config.label(), 1, 1, headerStyleID)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'application du style de header'")
	}

	for i, size := range config.sizes() {
		err := setColSize(xls, config.label(), i+1, size) // les colonnes en excel commencent à 1
		if err != nil {
			return errors.Wrap(err, "erreur pendant la mise à jour de la largeur de colonne")
		}
	}
	return nil
}

func setColSize(xls *excelize.File, sheetName string, col int, size float64) error {
	name, err := excelize.ColumnNumberToName(col)
	if err != nil {
		return errors.Wrap(err, "erreur pendant la récupération des coordonées")
	}
	err = xls.SetColWidth(sheetName, name, name, size)
	return err
}

func writeString(f *excelize.File, sheetName string, value string, col, row int) error {
	cell, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant la récupération des coordonnées")
	}
	err = f.SetCellStr(sheetName, cell, value)
	return err
}

func writeRow(f *excelize.File, sheetName string, values []any, row int) error {
	cell, err := excelize.CoordinatesToCellName(1, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant la récupération des coordonnées")
	}
	err = f.SetSheetRow(sheetName, cell, &values)
	if err != nil {
		return fmt.Errorf("erreur pendant l'écriture de la ligne : %w", err)
	}
	return nil
}

func writeInt(f *excelize.File, sheetName string, value int, col, row int) error {
	return writeString(f, sheetName, strconv.Itoa(value), col, row)
}

func exportTo(output io.Writer, xlsx *excelize.File) error {
	err := xlsx.Write(output, excelize.Options{RawCellValue: true})
	if err != nil {
		slog.Error("erreur pendant l'écriture du fichier xlsx")
		return err
	}
	return nil
}
