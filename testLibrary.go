package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/pmezard/go-difflib/difflib"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"testing"
)

var update, _ = strconv.ParseBool(os.Getenv("GOLDEN_UPDATE"))

func compare(expected []byte, actual []byte) string {
	diff := difflib.UnifiedDiff{
		A:       difflib.SplitLines(string(expected)),
		B:       difflib.SplitLines(string(actual)),
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

// not a test function
func followEntreprise(t *testing.T, siren string) {
	_, data, err := get(t, "/entreprise/get/"+siren)
	if err != nil {
		t.Fatalf("error when get entreprise with siren '%s' -> %s", siren, err)
	}
	entreprise := jsonToEntreprise(t, data)
	siret := entreprise.EtablissementsSummary[0].Siret
	t.Logf("will follow etablissement with siret '%s for entreprise with siren '%s'", siret, siren)
	followEtab(t, siret)
}

// not a test function
func followEtab(t *testing.T, siret string) {
	params := map[string]interface{}{
		"comment":  "test",
		"category": "test",
	}
	resp, _, _ := post(t, "/follow/"+siret, params)
	if resp.StatusCode != 201 {
		t.Errorf("le suivi a échoué: %d", resp.StatusCode)
	}
}

func jsonToEntreprise(t *testing.T, data []byte) Entreprise {
	var entreprise Entreprise
	err := json.Unmarshal(data, &entreprise)
	if err != nil {
		t.Fatalf("error when unmarshalling entreprise -> %s", err)
	}
	return entreprise
}

func hostname() string {
	return os.Getenv("DATAPI_URL")
}

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

// VAF encode le statut d'une entreprise selon la classification VAF
type VAF struct {
	visible  bool
	alert    bool
	followed bool
}

func razEtablissementFollowing(t *testing.T) {
	_, err := db.Exec(context.Background(), "delete from etablissement_follow;")
	if err != nil {
		t.Fatalf("Erreur d'accès lors du nettoyage pre-test de la base: %s", err.Error())
	}
}

// pgeTest struct contenant des données à tester pour le pge
type pgeTest struct {
	siren       string
	hasPGE      *bool
	mustFollow  bool
	expectedPGE *bool
}

// insertPGE add pge in entreprise_pge for a siren
func insertPGE(t *testing.T, pgesData []pgeTest) {
	tx, err := db.Begin(context.Background())
	if err != nil {
		t.Fatalf("something bad is happening with database when begining : %s" + err.Error())
	}

	for _, pgeTest := range pgesData {
		if pgeTest.hasPGE == nil {
			continue
		}
		_, err = tx.Exec(
			context.Background(),
			"insert into entreprise_pge (siren, actif) values ($1, $2)", pgeTest.siren, *pgeTest.hasPGE)
		if err != nil {
			tx.Rollback(context.Background())
			t.Fatalf("error inserting pge [%t] for siren '%s'. Cause : %s", *pgeTest.hasPGE, pgeTest.siren, err.Error())
		}
	}
	err = tx.Commit(context.Background())
	if err != nil {
		panic("something bad is happening with database" + err.Error())
	}
}

// selectSomeSiretsToFollow // récupérer une liste de sirets à suivre de toutes les typologies d'établissements
func selectSomeSiretsToFollow(t *testing.T) []string {
	sirets := getSiret(t, VAF{false, false, false}, 1)
	sirets = append(sirets, getSiret(t, VAF{false, true, false}, 1)...)
	sirets = append(sirets, getSiret(t, VAF{true, false, false}, 1)...)
	sirets = append(sirets, getSiret(t, VAF{true, true, false}, 1)...)
	t.Logf("sirests to follow -> %s", sirets)
	return sirets
}
