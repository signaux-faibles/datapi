package test

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/signaux-faibles/datapi/src/core"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
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

func ProcessGoldenFile(t *testing.T, path string, data []byte) (string, error) {
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

func Post(t *testing.T, path string, params map[string]interface{}) (*http.Response, []byte, error) {
	jsonValue, _ := json.Marshal(params)
	resp, err := http.Post(hostname()+path, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		t.Errorf("api non joignable: %s", err)
		return nil, nil, err
	}
	indented, err := indent(resp.Body)
	return resp, indented, err
}

func Get(t *testing.T, path string) (*http.Response, []byte, error) {
	resp, err := http.Get(hostname() + path)
	if err != nil {
		t.Errorf("api non joignable: %s", err)
		return nil, nil, err
	}
	indented, err := indent(resp.Body)
	return resp, indented, err
}

func FollowEntreprise(t *testing.T, siren string) {
	_, data, err := Get(t, "/entreprise/get/"+siren)
	if err != nil {
		t.Fatalf("error when get entreprise with siren '%s' -> %s", siren, err)
	}
	entreprise := JsonToEntreprise(t, data)
	siret := entreprise.EtablissementsSummary[0].Siret
	t.Logf("will follow etablissement with siret '%s for entreprise with siren '%s'", siret, siren)
	FollowEtablissement(t, siret)
}

func FollowEtablissement(t *testing.T, siret string) {
	params := map[string]interface{}{
		"comment":  "test",
		"category": "test",
	}
	resp, _, _ := Post(t, "/follow/"+siret, params)
	if resp.StatusCode != 201 {
		t.Errorf("le suivi a échoué: %d", resp.StatusCode)
	}
	t.Logf("nouvel établissement suivi -> '%s'", siret)
}

func JsonToEntreprise(t *testing.T, data []byte) core.Entreprise {
	var entreprise core.Entreprise
	err := json.Unmarshal(data, &entreprise)
	if err != nil {
		t.Fatalf("error when unmarshalling entreprise -> %s", err)
	}
	return entreprise
}

func hostname() string {
	return os.Getenv("DATAPI_URL")
}

func GetSiret(t *testing.T, v VAF, n int) []string {
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
	rows, err := core.Db().Query(
		context.Background(),
		sql,
		v.Visible,
		v.Alert,
		v.Followed,
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
	Visible  bool
	Alert    bool
	Followed bool
}

func RazEtablissementFollowing(t *testing.T) {
	_, err := core.Db().Exec(context.Background(), "delete from etablissement_follow;")
	if err != nil {
		t.Fatalf("Erreur d'accès lors du nettoyage pre-test de la base: %s", err.Error())
	}
	t.Log("informations de suivi des etablissements effacées")
}

// PgeTest struct contenant des données à tester pour le pge
type PgeTest struct {
	Siren       string
	HasPGE      *bool
	MustFollow  bool
	ExpectedPGE *bool
}

// InsertPGE add pge in entreprise_pge for a siren
func InsertPGE(t *testing.T, pgesData []PgeTest) {
	tx, err := core.Db().Begin(context.Background())
	if err != nil {
		t.Fatalf("something bad is happening with database when begining : %s" + err.Error())
	}

	for _, pgeTest := range pgesData {
		if pgeTest.HasPGE == nil {
			continue
		}
		_, err = tx.Exec(
			context.Background(),
			"insert into entreprise_pge (siren, actif) values ($1, $2)", pgeTest.Siren, *pgeTest.HasPGE)
		if err != nil {
			tx.Rollback(context.Background())
			t.Fatalf("error inserting pge [%t] for siren '%s'. Cause : %s", *pgeTest.HasPGE, pgeTest.Siren, err.Error())
		}
	}
	err = tx.Commit(context.Background())
	if err != nil {
		panic("something bad is happening with database" + err.Error())
	}
}

// SelectSomeSiretsToFollow // récupérer une liste de sirets à suivre de toutes les typologies d'établissements
func SelectSomeSiretsToFollow(t *testing.T) []string {
	sirets := GetSiret(t, VAF{false, false, false}, 1)
	sirets = append(sirets, GetSiret(t, VAF{false, true, false}, 1)...)
	sirets = append(sirets, GetSiret(t, VAF{true, false, false}, 1)...)
	sirets = append(sirets, GetSiret(t, VAF{true, true, false}, 1)...)
	t.Logf("sirets to follow -> %s", sirets)
	return sirets
}

func ExclureSuivi(t *testing.T) {
	var params = make(map[string]interface{})
	params["exclureSuivi"] = true
	params["ignoreZone"] = true
	_, indented, _ := Post(t, "/scores/liste", params)
	var liste Liste
	json.Unmarshal(indented, &liste)
	if len(liste.Scores) < 1 {
		t.Error("pas assez d'établissements, test faible")
	}

	siret := liste.Scores[0].Siret
	t.Logf("suivi de l'établissement %s", siret)
	resp, _, _ := Post(t, "/follow/"+siret, map[string]interface{}{
		"comment":  "test",
		"category": "test",
	})

	if resp.StatusCode != 201 {
		t.Errorf("le suivi a échoué: %d", resp.StatusCode)
	}

	_, indented, _ = Post(t, "/scores/liste", params)
	json.Unmarshal(indented, &liste)
	for _, l := range liste.Scores {
		if l.Followed {
			t.Fail()
		}
	}
}

// Liste de détection
type Liste struct {
	Scores []struct {
		Siret    string `json:"siret"`
		Followed bool   `json:"followed"`
	} `json:"scores"`
}

func (v *VAF) Read(vaf string) {
	v.Visible = vaf[0] == 'V'
	v.Alert = vaf[1] == 'A'
	v.Followed = vaf[2] == 'F'
}

type EtablissementVAF struct {
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

type SearchVAF struct {
	Results []EtablissementVAF `json:"results"`
}
