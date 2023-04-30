package core

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/signaux-faibles/datapi/src/db"
	"github.com/signaux-faibles/datapi/src/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/tealeg/xlsx"
)

type paramsListeScores struct {
	Departements          []string `json:"zone,omitempty"`
	EtatsProcol           []string `json:"procol,omitempty"`
	Activites             []string `json:"activite,omitempty"`
	EffectifMin           *int     `json:"effectifMin"`
	EffectifMax           *int     `json:"effectifMax"`
	EffectifMinEntreprise *int     `json:"effectifMinEntreprise"`
	EffectifMaxEntreprise *int     `json:"effectifMaxEntreprise"`
	CaMin                 *int     `json:"caMin"`
	CaMax                 *int     `json:"caMax"`
	IgnoreZone            *bool    `json:"ignorezone"`
	ExclureSuivi          *bool    `json:"exclureSuivi"`
	SiegeUniquement       bool     `json:"siegeUniquement"`
	Page                  int      `json:"page"`
	Filter                string   `json:"filter"`
	ExcludeSecteursCovid  []string `json:"excludeSecteursCovid"`
	EtatAdministratif     *string  `json:"etatAdministratif"`
	FirstAlert            *bool    `json:"firstAlert"`
}

// Liste de détection
type Liste struct {
	ID          string            `json:"id"`
	Batch       string            `json:"batch"`
	Algo        string            `json:"algo"`
	Description *string           `json:"description,omitempty"`
	Query       paramsListeScores `json:"-"`
	Scores      []*Summary        `json:"scores,omitempty"`
	NbF1        int               `json:"nbF1"`
	NbF2        int               `json:"nbF2"`
	From        int               `json:"from,omitempty"`
	Total       int               `json:"total"`
	Page        int               `json:"page,omitempty"`
	PageMax     int               `json:"pageMax,omitempty"`
	To          int               `json:"to,omitempty"`
	CurrentList bool              `json:"-"`
}

func getListes(c *gin.Context) {
	listes, err := findAllListes()
	if err != nil {
		c.JSON(500, err.Error())
	}
	c.JSON(200, listes)
}

func getLastListeScores(c *gin.Context) {
	roles := scopeFromContext(c)
	username := c.GetString("username")
	listes, err := findAllListes()
	if err != nil || len(listes) == 0 {
		c.AbortWithStatus(204)
		return
	}

	var params paramsListeScores
	err = c.Bind(&params)

	if params.EtatAdministratif != nil && !(*params.EtatAdministratif == "A" || *params.EtatAdministratif == "F") {
		c.JSON(400, "etatAdministratif must be either absent or `A` or `F`")
	}

	if err != nil || len(listes) == 0 {
		c.AbortWithStatus(400)
		return
	}
	liste := Liste{
		ID:          listes[0].ID,
		Query:       params,
		CurrentList: true,
	}
	limit := viper.GetInt("searchPageLength")
	if limit == 0 {
		c.JSON(418, "searchPageLength must be > 0 in configuration therefore, I'm a teapot.")
		return
	}
	Jerr := liste.getScores(roles, params.Page, &limit, username)
	if Jerr != nil {
		c.JSON(Jerr.Code(), Jerr.Error())
		return
	}
	c.JSON(200, liste)
}

func getXLSListeScores(c *gin.Context) {
	roles := scopeFromContext(c)
	username := c.GetString("username")

	var params paramsListeScores
	err := c.Bind(&params)

	listes, err := findAllListes()
	if err != nil || len(listes) == 0 {
		c.AbortWithStatus(204)
		return
	}

	liste := Liste{
		ID:          c.Param("id"),
		Query:       params,
		CurrentList: listes[0].ID == c.Param("id"),
	}

	if err != nil || liste.ID == "" {
		c.AbortWithStatus(400)
		return
	}
	limit := viper.GetInt("searchPageLength")
	if limit == 0 {
		c.JSON(418, "searchPageLength must be > 0 in configuration therefore, I'm a teapot.")
		return
	}
	Jerr := liste.getScores(roles, 0, nil, username)

	if Jerr != nil {
		c.JSON(Jerr.Code(), Jerr.Error())
		return
	}

	file, err := liste.toXLS(params)
	c.Writer.Header().Set("Content-disposition", "attachment;filename=extract.xls")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", file)
}

func getListeScores(c *gin.Context) {
	roles := scopeFromContext(c)
	username := c.GetString("username")

	var params paramsListeScores
	err := c.Bind(&params)
	if err != nil {
		c.AbortWithStatus(400)
		return
	}

	if params.EtatAdministratif != nil {
		if !(*params.EtatAdministratif == "A" || *params.EtatAdministratif == "F") {
			c.JSON(400, "etatAdministratif must be either absent or 'A' or 'F'")
			return
		}
	}

	listes, err := findAllListes()
	if err != nil || len(listes) == 0 {
		c.AbortWithStatus(204)
		return
	}

	liste := Liste{
		ID:          c.Param("id"),
		Query:       params,
		CurrentList: listes[0].ID == c.Param("id"),
	}

	if liste.ID == "" {
		c.AbortWithStatus(400)
		return
	}

	limit := viper.GetInt("searchPageLength")
	if limit == 0 {
		c.JSON(418, "searchPageLength must be > 0 in configuration therefore, I'm a teapot.")
		return
	}
	Jerr := liste.getScores(roles, params.Page, &limit, username)

	if Jerr != nil {
		c.JSON(Jerr.Code(), Jerr.Error())
		return
	}

	c.JSON(200, liste)
}

func (liste *Liste) getScores(roles Scope, page int, limit *int, username string) utils.Jerror {
	if liste.Batch == "" {
		err := liste.load()
		if err != nil {
			return err
		}
	}
	var offset int
	if limit == nil {
		offset = 0
	} else {
		offset = page * *limit
	}
	var suivi *bool
	if liste.Query.ExclureSuivi != nil {
		if *liste.Query.ExclureSuivi {
			s := false
			suivi = &s
		}
	}

	params := summaryParams{
		roles, limit, &offset, &liste.ID, liste.CurrentList, &liste.Query.Filter, nil,
		liste.Query.IgnoreZone, username, liste.Query.SiegeUniquement, "score", &True, liste.Query.EtatsProcol,
		liste.Query.Departements, suivi, liste.Query.EffectifMin, liste.Query.EffectifMax, nil, liste.Query.Activites,
		liste.Query.EffectifMinEntreprise, liste.Query.EffectifMaxEntreprise, liste.Query.CaMin, liste.Query.CaMax,
		liste.Query.ExcludeSecteursCovid, liste.Query.EtatAdministratif, liste.Query.FirstAlert,
	}
	summaries, err := getSummaries(params)
	if err != nil {
		return utils.ErrorToJSON(http.StatusInternalServerError, err)
	}

	scores := summaries.Summaries
	// TODO: harmoniser les types de sorties pour éviter les remaniements
	if summaries.Global.Count != nil {
		liste.Total = *summaries.Global.Count
		liste.NbF1 = *summaries.Global.CountF1
		liste.NbF2 = *summaries.Global.CountF2
	}

	if limit == nil {
		i := 1
		limit = &i
	}
	liste.Scores = scores
	liste.Page = page
	liste.PageMax = (liste.Total - 1) / *limit
	liste.From = *limit*page + 1
	liste.To = *limit*page + len(scores)
	if *limit*page > liste.Total || liste.Total == 0 {
		return utils.NewJSONerror(http.StatusNoContent, "empty page")
	}
	return nil
}

func findAllListes() ([]Liste, error) {
	var listes []Liste
	rows, err := db.Get().Query(context.Background(), `
		select algo, batch, libelle, description from liste l
		left join liste_description d on d.libelle_liste = l.libelle
		where version=0 order by batch desc, algo
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var l Liste
		err := rows.Scan(&l.Algo, &l.Batch, &l.ID, &l.Description)
		if err != nil {
			return nil, err
		}
		listes = append(listes, l)
	}

	if len(listes) == 0 {
		return nil, errors.New("findAllListes: no list found")
	}
	return listes, nil
}

func (liste *Liste) load() utils.Jerror {
	sqlListe := `select batch, algo from liste where libelle=$1 and version=0`
	row := db.Get().QueryRow(context.Background(), sqlListe, liste.ID)
	batch, algo := "", ""
	err := row.Scan(&batch, &algo)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return utils.NewJSONerror(404, "no such list")
		}
		return utils.ErrorToJSON(500, err)
	}
	liste.Batch = batch
	liste.Algo = algo
	return nil
}

func (liste *Liste) toXLS(params paramsListeScores) ([]byte, utils.Jerror) {
	xlFile := xlsx.NewFile()
	xlSheet, err := xlFile.AddSheet("extract")
	if err != nil {
		return nil, utils.ErrorToJSON(500, err)
	}

	row := xlSheet.AddRow()
	row.AddCell().Value = "liste"
	row.AddCell().Value = "siren"
	row.AddCell().Value = "siret"
	row.AddCell().Value = "departement"
	row.AddCell().Value = "raison_sociale"
	row.AddCell().Value = "dernier_effectif_entreprise"
	row.AddCell().Value = "code_activite"
	row.AddCell().Value = "libelle_activite"
	row.AddCell().Value = "alert"
	row.AddCell().Value = "premiere_alerte"
	row.AddCell().Value = "secteur_covid"
	row.AddCell().Value = "chiffre_affaire"
	row.AddCell().Value = "excedent_brut_exploitation"
	row.AddCell().Value = "resultat_d_exploitation"
	row.AddCell().Value = "date_cloture_bilan"
	row.AddCell().Value = "etat_procol"
	row.AddCell().Value = "tete_groupe"

	for _, score := range liste.Scores {
		row := xlSheet.AddRow()
		if err != nil {
			return nil, utils.ErrorToJSON(500, err)
		}
		row.AddCell().Value = liste.ID
		row.AddCell().Value = score.Siren
		row.AddCell().Value = score.Siret
		row.AddCell().Value = *score.CodeDepartement
		row.AddCell().Value = *score.RaisonSociale

		if score.EffectifEntreprise != nil {
			cell := row.AddCell()
			cell.SetInt(int(*score.EffectifEntreprise))
		} else {
			row.AddCell().Value = "n/c"
		}

		row.AddCell().Value = *score.CodeActivite
		if score.LibelleActivite != nil {
			row.AddCell().Value = *score.LibelleActivite
		} else {
			row.AddCell().Value = "n/c"
		}

		if score.Alert != nil {
			if *score.Alert == "Alerte seuil F1" {
				row.AddCell().Value = "Risque élevé"
			}
			if *score.Alert == "Alerte seuil F2" {
				row.AddCell().Value = "Risque modéré"
			}
			if *score.Alert == "Pas d'alerte" {
				row.AddCell().Value = "Pas de risque"
			}
		} else {
			row.AddCell().Value = "Hors périmètre"
		}

		if *score.FirstAlert {
			row.AddCell().Value = "oui"
		} else {
			row.AddCell().Value = "non"
		}
		row.AddCell().Value = *score.SecteurCovid
		if score.ChiffreAffaire != nil {
			row.AddCell().Value = fmt.Sprintf("%d k€", int(*score.ChiffreAffaire))
		} else {
			row.AddCell().Value = "n/c"
		}
		if score.ExcedentBrutDExploitation != nil {
			row.AddCell().Value = fmt.Sprintf("%d k€", int(*score.ExcedentBrutDExploitation))
		} else {
			row.AddCell().Value = "n/c"
		}
		if score.ResultatExploitation != nil {
			row.AddCell().Value = fmt.Sprintf("%d k€", int(*score.ResultatExploitation))
		} else {
			row.AddCell().Value = "n/c"
		}
		if score.ArreteBilan != nil {
			row.AddCell().Value = score.ArreteBilan.Format("02/01/2006")
		} else {
			row.AddCell().Value = "n/c"
		}
		if score.EtatProcol != nil {
			row.AddCell().Value = *score.EtatProcol
		} else {
			row.AddCell().Value = "n/c"
		}
		if score.Groupe != nil {
			row.AddCell().Value = *score.Groupe
		} else {
			row.AddCell().Value = "pas de groupe identifié"
		}
	}

	sheetParams, _ := xlFile.AddSheet("parameters")
	row = sheetParams.AddRow()
	s, _ := json.Marshal(params.Activites)
	row.AddCell().Value = "activites"
	row.AddCell().Value = string(s)

	row = sheetParams.AddRow()
	s, _ = json.Marshal(params.Departements)
	row.AddCell().Value = "departements"
	row.AddCell().Value = string(s)

	row = sheetParams.AddRow()
	s, _ = json.Marshal(params.EffectifMin)
	row.AddCell().Value = "effectifMin"
	row.AddCell().Value = string(s)

	row = sheetParams.AddRow()
	s, _ = json.Marshal(params.EffectifMax)
	row.AddCell().Value = "effectifMax"
	row.AddCell().Value = string(s)

	row = sheetParams.AddRow()
	s, _ = json.Marshal(params.EtatsProcol)
	row.AddCell().Value = "etatsProcol"
	row.AddCell().Value = string(s)

	row = sheetParams.AddRow()
	s, _ = json.Marshal(params.Filter)
	row.AddCell().Value = "filter"
	row.AddCell().Value = string(s)

	row = sheetParams.AddRow()
	s, _ = json.Marshal(params.IgnoreZone)
	row.AddCell().Value = "ignorezone"
	row.AddCell().Value = string(s)

	data := bytes.NewBuffer(nil)
	file := bufio.NewWriter(data)
	xlFile.Write(file)
	file.Flush()
	return data.Bytes(), nil
}
