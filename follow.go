package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Follow type follow pour l'API
type Follow struct {
	Siret                *string   `json:"siret,omitempty"`
	Username             *string   `json:"username,omitempty"`
	Active               bool      `json:"active"`
	Since                time.Time `json:"since"`
	Comment              string    `json:"comment"`
	Category             string    `json:"category"`
	UnfollowComment      string    `json:"unfollowComment,omitempty"`
	UnfollowCategory     string    `json:"unfollowCategory,omitempty"`
	EtablissementSummary *Summary  `json:"etablissementSummary,omitempty"`
}

func getEtablissementsFollowedByCurrentUser(c *gin.Context) {
	username := c.GetString("username")
	scope := scopeFromContext(c)
	follow := Follow{Username: &username}
	follows, err := follow.list(scope)

	if err != nil {
		c.JSON(err.Code(), err.Error())
		return
	}

	c.JSON(200, follows)
}

func getXLSXFollowedByCurrentUser(c *gin.Context) {
	username := c.GetString("username")
	scope := scopeFromContext(c)

	wekan := contains(scope, "wekan")
	export, err := getExport(scope, username, wekan, nil)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	xlsx, err := export.xlsx(wekan)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	filename := fmt.Sprintf("export-suivi-%s.xlsx", time.Now().Format("060102"))
	c.Writer.Header().Set("Content-disposition", "attachment;filename="+filename)
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", xlsx)
}

func getDOCXFollowedByCurrentUser(c *gin.Context) {
	username := c.GetString("username")
	scope := scopeFromContext(c)
	auteur := c.GetString("given_name") + " " + c.GetString("family_name")
	wekan := contains(scope, "wekan")
	export, err := getExport(scope, username, wekan, nil)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	header := ExportHeader{
		Auteur: auteur,
		Date:   time.Now(),
	}
	docx, err := export.docx(header)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	filename := fmt.Sprintf("export-suivi-%s.docx", time.Now().Format("060102"))
	c.Writer.Header().Set("Content-disposition", "attachment;filename="+filename)
	c.Data(200, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", docx)
}

func getDOCXFromSiret(c *gin.Context) {
	username := c.GetString("username")
	auteur := c.GetString("given_name") + " " + c.GetString("family_name")
	scope := scopeFromContext(c)
	sirets := append([]string{}, c.Param("siret"))
	wekan := contains(scope, "wekan")
	export, err := getExport(scope, username, wekan, sirets)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	header := ExportHeader{
		Auteur: auteur,
		Date:   time.Now(),
	}
	docx, err := export.docx(header)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	filename := fmt.Sprintf("export-%s-%s.docx", sirets[0], time.Now().Format("060102"))
	c.Writer.Header().Set("Content-disposition", "attachment;filename="+filename)
	c.Data(200, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", docx)
}

func followEtablissement(c *gin.Context) {
	siret := c.Param("siret")
	username := c.GetString("username")

	var param struct {
		Comment  string `json:"comment"`
		Category string `json:"category"`
	}
	paramErr := c.ShouldBind(&param)
	if paramErr != nil {
		c.AbortWithError(500, paramErr)
		return
	}
	if param.Category == "" {
		c.JSON(400, "mandatory non-empty `category` property")
		return
	}

	follow := Follow{
		Siret:    &siret,
		Username: &username,
		Comment:  param.Comment,
		Category: param.Category,
	}

	err := follow.load()
	if err != nil && err.Error() != "no rows in result set" {
		c.AbortWithError(500, err)
		return
	}

	if follow.Active {
		c.JSON(204, follow)
	} else {
		err := follow.activate()
		if err != nil && err.Error() != "no rows in result set" {
			c.AbortWithError(500, err)
			return
		} else if err != nil && err.Error() == "no rows in result set" {
			c.JSON(403, "unknown establishment")
			return
		}
		c.JSON(201, follow)
	}
}

func unfollowEtablissement(c *gin.Context) {
	siret := c.Param("siret")
	username := c.GetString("username")

	var param struct {
		UnfollowComment  string `json:"unfollowComment"`
		UnfollowCategory string `json:"unfollowCategory"`
	}
	paramErr := c.ShouldBind(&param)
	if paramErr != nil {
		c.AbortWithError(500, paramErr)
		return
	}
	if param.UnfollowCategory == "" {
		c.JSON(400, "mandatory non-empty `unfollowCategory` property")
		return
	}

	follow := Follow{
		Siret:            &siret,
		Username:         &username,
		UnfollowComment:  param.UnfollowComment,
		UnfollowCategory: param.UnfollowCategory,
	}

	err := follow.deactivate()
	if err != nil {
		c.JSON(err.Code(), err.Error())
		return
	}
	c.JSON(200, "this establishment is no longer followed")
}

func (f *Follow) load() error {
	sqlFollow := `select 
        active, since, comment, category        
        from etablissement_follow 
        where
        username = $1 and
        siret = $2 and
        active`

	return db.QueryRow(context.Background(), sqlFollow, f.Username, f.Siret).Scan(&f.Active, &f.Since, &f.Comment, &f.Category)
}

func (f *Follow) activate() error {
	f.Active = true
	sqlActivate := `insert into etablissement_follow
    (siret, siren, username, active, since, comment, category)
    select $1, $2, $3, true, current_timestamp, $4, $5
    from etablissement0 e
    where siret = $6
    returning since, true`

	return db.QueryRow(context.Background(),
		sqlActivate,
		*f.Siret,
		(*f.Siret)[0:9],
		f.Username,
		f.Comment,
		f.Category,
		f.Siret,
	).Scan(&f.Since, &f.Active)
}

func (f *Follow) deactivate() Jerror {
	sqlUnactivate := `update etablissement_follow set active = false, until = current_timestamp, unfollow_comment = $3, unfollow_category = $4
        where siret = $1 and username = $2 and active = true`

	commandTag, err := db.Exec(context.Background(),
		sqlUnactivate, f.Siret, f.Username, f.UnfollowComment, f.UnfollowCategory)

	if err != nil {
		return errorToJSON(500, err)
	}

	if commandTag.RowsAffected() == 0 {
		return newJSONerror(204, "this establishment is already not followed")
	}

	return nil
}

func (f *Follow) list(roles scope) (Follows, Jerror) {
	liste, err := findAllListes()
	if err != nil {
		return nil, errorToJSON(500, err)
	}

	params := summaryParams{roles.zoneGeo(), nil, nil, &liste[0].ID, false, nil,
		&True, &True, *f.Username, false, "follow", &False, nil,
		nil, &True, nil, nil, nil, nil, nil, nil, nil, nil}

	sms, err := getSummaries(params)
	if err != nil {
		return nil, errorToJSON(500, err)
	}
	var follows Follows

	for _, s := range sms.summaries {
		var f Follow
		f.Comment = *coalescepString(s.Comment, &EmptyString)
		f.Category = *coalescepString(s.Category, &EmptyString)
		f.Since = *coalescepTime(s.Since, &time.Time{})
		f.EtablissementSummary = s
		f.Active = true
		follows = append(follows, f)
	}

	if len(follows) == 0 {
		return nil, newJSONerror(204, "aucun suivi")
	}

	return follows, nil
}

// Follows is a slice of Follows
type Follows []Follow
