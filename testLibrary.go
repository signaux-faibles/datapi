package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/pmezard/go-difflib/difflib"
)

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
	diff, err := compare(goldenFile, data), nil
	if diff != "" {
		t.Errorf("differences entre le résultat et le golden file: %s \n%s", path, diff)
	}
	return diff, err
}

func post(t *testing.T, path string, params map[string]interface{}) (*http.Response, []byte, error) {
	jsonValue, _ := json.Marshal(params)
	resp, err := http.Post(hostname()+path, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		t.Errorf("api non joignable: %s", err)
		return nil, nil, err
	}
	indented, err := indent(resp.Body)
	return resp, indented, err
}

func get(t *testing.T, path string) (*http.Response, []byte, error) {
	resp, err := http.Get(hostname() + path)
	if err != nil {
		t.Errorf("api non joignable: %s", err)
		return nil, nil, err
	}
	indented, err := indent(resp.Body)
	return resp, indented, err
}

//func connect() *pgxpool.Pool {
//	viper.SetConfigName("config")
//	viper.SetConfigType("toml")
//	viper.AddConfigPath("workspace")
//	viper.ReadInConfig()
//	pgConnStr := viper.GetString("postgres")
//	db, err := pgxpool.Connect(context.Background(), pgConnStr)
//	if err != nil {
//		log.Fatalf("database connexion: %s", err.Error())
//	}
//	// Suppression des éventuels suivi d'un test précédent
//	_, err = db.Exec(context.Background(), `delete from etablissement_follow;`)
//	if err != nil {
//		log.Fatalf("erreur lors de la suppression des suivis: %s", err.Error())
//	}
//
//	return db
//}

func getSiret(t *testing.T, v VAF, n int) []string {
	sql := `with etablissement_data_confidentielle as (
		select distinct e.siret, et.departement
		from etablissement_periode_urssaf e
		inner join etablissement et on et.siret = e.siret
		inner join etablissement_apdemande ap on ap.siret = e.siret
		where part_patronale+part_salariale != 0
	)
	select e.siret from etablissement e
	inner join etablissement_data_confidentielle d on d.siret = e.siret
	left join v_roles r on r.siren = e.siren and roles && array['75','77','78','91','92','93','94']
	left join v_entreprise_follow f on f.siren = e.siren
	left join v_alert_entreprise a on a.siren = e.siren
	where 
		e.departement != '20' and
		(r.siren is not null)=$1 and -- roles
		(a.siren is not null)=$2 and -- alert
		(f.siren is not null)=$3     -- follow
	order by e.siret
	limit $4`
	rows, err := db.Query(
		context.Background(),
		sql,
		v.visible,
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

func testEtablissementVAF(t *testing.T, siret string, vaf string) {
	v := VAF{}
	v.read(vaf)

	goldenFilePath := fmt.Sprintf("data/getEtablissement-%s-%s.json.gz", vaf, siret)
	t.Logf("l'établissement %s est bien de la forme attendue (ref %s)", siret, goldenFilePath)
	_, indented, _ := get(t, "/etablissement/get/"+siret)
	processGoldenFile(t, goldenFilePath, indented)

	var e etablissementVAF
	json.Unmarshal(indented, &e)
	if !(e.Visible == v.visible && e.Followed == v.followed) {
		t.Errorf("l'établissement %s de type %s n'a pas les propriétés requises", siret, vaf)
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

func testSearchVAF(t *testing.T, siret string, vaf string) {
	goldenFilePath := fmt.Sprintf("data/getSearch-%s-%s.json.gz", vaf, siret)
	t.Logf("la recherche renvoie l'établissement %s sous la forme attendue (ref %s)", siret, goldenFilePath)
	params := map[string]interface{}{
		"search":      siret,
		"ignoreZone":  true,
		"ignoreRoles": true,
	}
	_, indented, _ := post(t, "/etablissement/search", params)
	diff, _ := processGoldenFile(t, goldenFilePath, indented)
	if diff != "" {
		t.Errorf("differences entre le résultat et le golden file: %s \n%s", goldenFilePath, diff)
	}
	visible := vaf[0] == 'V'
	followed := vaf[2] == 'F'
	var e searchVAF
	json.Unmarshal(indented, &e)
	if len(e.Results) != 1 || !(e.Results[0].Visible == visible && e.Results[0].Followed == followed) {
		fmt.Println(vaf, visible, followed)
		t.Errorf("la recherche %s de type %s n'a pas les propriétés requises", siret, vaf)
	}
}

type searchVAF struct {
	Results []etablissementVAF `json:"results"`
}

// VAF encode le statut d'une entreprise selon la classification VAF
type VAF struct {
	visible  bool
	alert    bool
	followed bool
}

func (v *VAF) read(vaf string) {
	v.visible = vaf[0] == 'V'
	v.alert = vaf[1] == 'A'
	v.followed = vaf[2] == 'F'
}

type etablissementVAF struct {
	Visible       bool `json:"visible"`
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

func hostname() string {
	return os.Getenv("DATAPI_URL")
}
