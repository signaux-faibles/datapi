package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
)

//func loadCardsFromFile(cards *WekanCards, path string) error {
//	fileContent, err := os.ReadFile(path)
//	if err != nil {
//		return err
//	}
//	err = json.Unmarshal(fileContent, cards)
//	return err
//}

//func loadFollowExportsFromFile(exports *KanbanDBExports, path string) error {
//	fileContent, err := os.ReadFile(path)
//	if err != nil {
//		return err
//	}
//	err = json.Unmarshal(fileContent, exports)
//	return err
//}

func Test_WekanExportsDOCX(t *testing.T) {
	t.Log("KanbanExports can generate a non-zero length docx file")
	cards := KanbanExports{}

	docxifyPath, _ := filepath.Abs("../../build-container/docxify3.py")
	docxifyWorkingDir, _ := filepath.Abs("../../build-container")

	viper.Set("docxifyPath", docxifyPath)
	viper.Set("docxifyWorkingDir", docxifyWorkingDir)
	viper.Set("docxifyPython", "python3")
	dateHeader, _ := time.Parse("02/01/2006", "05/06/2018")
	header := ExportHeader{
		Auteur: "test_auteur",
		Date:   dateHeader,
	}
	var docxs Docxs
	for _, card := range cards {
		docx, err := card.docx(header)
		if err != nil {
			t.Logf("Error -> %s", err.Error())
			t.Fail()
		}
		docxs = append(docxs, docx)
	}

	data := docxs.zip()
	if len(data) == 0 {
		t.Error("empty docx file returned")
	}
	if os.Getenv("WRITE_DOCX") == "true" {
		err := os.WriteFile("test_output.docx", data, 0755)
		if err != nil {
			t.Errorf("could create result file: %s", err.Error())
		}
	}
}

func Test_WekanExportsXLSX(t *testing.T) {
	t.Log("KanbanExports can generate a non-zero length xlsx file")
	wekanExports := KanbanExports{}

	xlsx, err := wekanExports.xlsx(true)
	if len(xlsx) == 0 {
		t.Error("empty xlsx file returned")
	}
	if err != nil {
		t.Error(err)
	}
}
