package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
)

var hostname = "http://localhost:" + os.Getenv("DATAPI_PORT")
var update, _ = strconv.ParseBool(os.Getenv("GOLDEN_UPDATE"))

func compare(golden []byte, result []byte) string {
	diff := difflib.UnifiedDiff{
		A:       difflib.SplitLines(string(golden)),
		B:       difflib.SplitLines(string(result)),
		Context: 1,
	}

	text, _ := difflib.GetUnifiedDiffString(diff)
	return text
}

func loadGoldenFile(file string) ([]byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(gzipReader)
}

func saveGoldenFile(fileName string, goldenData []byte) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	gzipWriter := gzip.NewWriter(f)
	defer gzipWriter.Close()
	_, err = gzipWriter.Write(goldenData)
	if err != nil {
		return err
	}
	err = gzipWriter.Flush()
	return err
}

func indent(reader io.Reader) ([]byte, error) {
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	var prettyBody bytes.Buffer
	err = json.Indent(&prettyBody, body, "", "  ")
	return prettyBody.Bytes(), err
}

func processGoldenFile(t *testing.T, path string, data []byte) (string, error) {
	if update {
		err := saveGoldenFile(path, data)
		if err != nil {
			return "", err
		}
		t.Logf("mise à jour du golden file %s", path)
		return "", nil
	}
	goldenFile, err := loadGoldenFile(path)
	if err != nil {
		t.Errorf("golden file non disponible: %s", err.Error())
		return "", err
	}
	return compare(goldenFile, data), nil
}

func post(t *testing.T, path string, params map[string]interface{}) (*http.Response, []byte, error) {
	jsonValue, _ := json.Marshal(params)
	resp, err := http.Post(hostname+path, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		t.Errorf("api non joignable: %s", err)
		return nil, nil, err
	}
	indented, err := indent(resp.Body)
	return resp, indented, err
}

func get(t *testing.T, path string) (*http.Response, []byte, error) {
	resp, err := http.Get(hostname + path)
	if err != nil {
		t.Errorf("api non joignable: %s", err)
		return nil, nil, err
	}
	indented, err := indent(resp.Body)
	return resp, indented, err
}

func TestListes(t *testing.T) {
	_, indented, _ := get(t, "/listes")
	diff, _ := processGoldenFile(t, "data/listes.json.gz", indented)
	if diff != "" {
		t.Errorf("differences entre le résultat et le golden file: \n%s", diff)
	}
}

func TestScores(t *testing.T) {
	t.Log("/scores/liste retourne le même résultat qu'attendu")
	_, indented, _ := post(t, "/scores/liste", nil)
	diff, _ := processGoldenFile(t, "data/scores.json.gz", indented)
	if diff != "" {
		t.Errorf("differences entre le résultat et le golden file: \n%s", diff)
	}

	t.Log("/scores/liste retourne le même résultat qu'attendu avec ignoreZone=true")
	params := map[string]interface{}{
		"ignoreZone": true,
	}
	_, indented, _ = post(t, "/scores/liste", params)
	diff, _ = processGoldenFile(t, "data/scores-ignoreZone.json.gz", indented)
	if diff != "" {
		t.Errorf("differences entre le résultat et le golden file: \n%s", diff)
	}
}
