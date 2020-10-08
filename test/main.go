package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/spf13/viper"
)

var hostname = "http://localhost:" + os.Getenv("DATAPI_PORT")
var update, _ = strconv.ParseBool(os.Getenv("GOLDEN_UPDATE"))
var db = connect()

func main() {
	fmt.Println("Just run go test")
}

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
	diff, err := compare(goldenFile, data), nil
	if diff != "" {
		t.Errorf("differences entre le résultat et le golden file: %s \n%s", path, diff)
	}
	return diff, err
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

func connect() *pgxpool.Pool {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("workspace")
	viper.ReadInConfig()
	pgConnStr := viper.GetString("postgres")
	db, err := pgxpool.Connect(context.Background(), pgConnStr)
	if err != nil {
		log.Fatalf("database connexion: %s", err.Error())
	}
	// Suppression des éventuels suivi d'un test précédent
	_, err = db.Exec(context.Background(), `delete from etablissement_follow;`)
	if err != nil {
		log.Fatalf("erreur lors de la suppression des suivis: %s", err.Error())
	}

	return db
}

func getSiret(t *testing.T, v VIAF, n int) []string {
	sql := `with etablissement_follow as (
		select coalesce(array_agg(distinct siret), '{}'::text[]) as f from etablissement_follow
	),
	etablissement_inzone as (
		select array_agg(distinct siret) as i from etablissement where
		departement in ('70','39','58','71','89','21','90','25','69','73','43','38','26','07','15','63','01','74','42','03')
	),
	etablissement_visible as (
		select array_agg(siret) as v from etablissement0 e
		inner join v_roles r on r.siren = e.siren
		where roles && array['70','39','58','71','89','21','90','25','69','73','43','38','26','07','15','63','01','74','42','03']
	),
	etablissement_alert as (
		select array_agg(siret) as a from etablissement0 e
		inner join v_alert_entreprise a on a.siren = e.siren
	),
	etablissement_data_confidentielle as (
		select distinct e.siret, et.departement
		from etablissement_periode_urssaf e
		inner join etablissement et on et.siret = e.siret
		inner join etablissement_apdemande ap on ap.siret = e.siret
		where part_patronale+part_salariale != 0 
	),
	etablissement_bool as (select e.siret,
		e.siret = any(v.v) as visible,
		e.siret = any(i.i) as inzone,
		e.siret = any(a.a) as alert,
		e.siret = any(f.f) as follow
	from etablissement_data_confidentielle e
	inner join etablissement_follow f on true
	inner join etablissement_inzone i on true
	inner join etablissement_visible v on true
	inner join etablissement_alert a on true
	where e.departement != '20')
	select siret from etablissement_bool
	where visible=$1 and inzone=$2 and alert=$3 and follow=$4
	order by siret
	limit $5`
	rows, err := db.Query(
		context.Background(),
		sql,
		v.visible,
		v.inZone,
		v.alert,
		v.followed,
		n,
	)
	if err != nil {
		t.Errorf("problème d'accès à la base de données: %s", err.Error())
		return nil
	}
	var sirets []string
	for rows.Next() {
		var siret string
		err := rows.Scan(&siret)
		if err != nil {
			t.Errorf("problème d'accès à la base de données: %s", err.Error())
			return nil
		}
		sirets = append(sirets, siret)
	}
	return sirets
}

func testEtablissementVIAF(t *testing.T, siret string, viaf string) {
	v := VIAF{}
	v.read(viaf)

	goldenFilePath := fmt.Sprintf("data/getEtablissement-%s-%s.json.gz", viaf, siret)
	t.Logf("l'établissement %s est bien de la forme attendue (ref %s)", siret, goldenFilePath)
	_, indented, _ := get(t, "/etablissement/get/"+siret)
	processGoldenFile(t, goldenFilePath, indented)

	var e etablissementVIAF
	json.Unmarshal(indented, &e)
	if !(e.Visible == v.visible && e.InZone == v.inZone && e.Followed == v.followed) {
		t.Errorf("l'établissement %s de type %s n'a pas les propriétés requises", siret, viaf)
	}

	if !((v.visible && v.alert) || v.followed) {
		if len(e.PeriodeUrssaf.Cotisation) > 0 {
			t.Errorf("Fuite de cotisations sur l'établissement %s (ref %s)", siret, goldenFilePath)
		}
		if len(e.PeriodeUrssaf.PartPatronale) > 0 {
			t.Errorf("Fuite de part patronale sur l'établissement %s (ref %s)", siret, goldenFilePath)
		}
		if len(e.PeriodeUrssaf.PartSalariale) > 0 {
			t.Errorf("Fuite de part salariale sur l'établissement %s (ref %s)", siret, goldenFilePath)
		}
		if len(e.Delai) > 0 {
			t.Errorf("Fuite de délai urssaf sur l'établissement %s (ref %s)", siret, goldenFilePath)
		}
		if len(e.APConso) > 0 {
			t.Errorf("Fuite de consommation d'activité partielle sur l'établissement %s (ref %s)", siret, goldenFilePath)
		}
		if len(e.APDemande) > 0 {
			t.Errorf("Fuite de demande d'activité partielle sur l'établissement %s (ref %s)", siret, goldenFilePath)
		}
	} else {
		if len(e.PeriodeUrssaf.Cotisation) == 0 {
			t.Errorf("Absence de cotisations urssaf sur l'établissement %s (ref %s)", siret, goldenFilePath)
		}
		if len(e.PeriodeUrssaf.PartPatronale) == 0 {
			t.Errorf("Absence de part patronale urssaf sur l'établissement %s (ref %s)", siret, goldenFilePath)
		}
		if len(e.PeriodeUrssaf.PartSalariale) == 0 {
			t.Errorf("Absence de part salariale urssaf sur l'établissement %s (ref %s)", siret, goldenFilePath)
		}
	}
}

func testSearchVIAF(t *testing.T, siret string, viaf string) {
	goldenFilePath := fmt.Sprintf("data/getSearch-%s-%s.json.gz", viaf, siret)
	t.Logf("la recherche renvoie l'établissement %s sous la forme attendue (ref %s)", siret, goldenFilePath)
	_, indented, _ := get(t, "/etablissement/search/"+siret+"?ignorezone=true&ignoreroles=true")
	diff, _ := processGoldenFile(t, goldenFilePath, indented)
	if diff != "" {
		t.Errorf("differences entre le résultat et le golden file: %s \n%s", goldenFilePath, diff)
	}
	visible := viaf[0] == 'V'
	inZone := viaf[1] == 'I'
	followed := viaf[3] == 'F'
	var e searchVIAF
	json.Unmarshal(indented, &e)
	if len(e.Results) != 1 || !(e.Results[0].Visible == visible && e.Results[0].InZone == inZone && e.Results[0].Followed == followed) {
		t.Errorf("la recherche %s de type %s n'a pas les propriétés requises", siret, viaf)
	}
}

type searchVIAF struct {
	Results []etablissementVIAF `json:"results"`
}

// VIAF hyperbool
type VIAF struct {
	visible  bool
	inZone   bool
	alert    bool
	followed bool
}

func (v *VIAF) read(viaf string) {
	v.visible = viaf[0] == 'V'
	v.inZone = viaf[1] == 'I'
	v.alert = viaf[2] == 'A'
	v.followed = viaf[3] == 'F'
}

type etablissementVIAF struct {
	Visible       bool `json:"visible"`
	InZone        bool `json:"inZone"`
	Alert         bool `json:"alert"`
	Followed      bool `json:"followed"`
	PeriodeUrssaf struct {
		Periodes      []*time.Time `json:"periodes"`
		Cotisation    []*float64   `json:"cotisation"`
		PartPatronale []*float64   `json:"partPatronale"`
		PartSalariale []*float64   `json:"partSalariale"`
		Effectif      []*int       `json:"effectif"`
	}
	APConso   []interface{} `json:"apConso"`
	APDemande []interface{} `json:"apDemande"`
	Delai     []interface{} `json:"delai"`
}
