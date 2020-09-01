package main

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Comment commentaire sur une enterprise
type Comment struct {
	ID             *int        `json:"id,omitempty"`
	IDParent       *int        `json:"idParent,omitempty"`
	Comments       []*Comment  `json:"comments,omitempty"`
	Siret          *string     `json:"siret,omitempty"`
	Username       *string     `json:"username,omitempty"`
	User           *user       `json:"author,omitempty"`
	DateHistory    []time.Time `json:"dateHistory,omitempty"`
	Message        *string     `json:"message,omitempty"`
	MessageHistory []string    `json:"messageHistory,omitempty"`
	Scope          [][]string  `json:"scope,omitempty"`
}

func getEntrepriseComments(c *gin.Context) {
	siret := c.Param("siret")
	comment := Comment{Siret: &siret}
	err := comment.load()
	if err != nil {
		c.JSON(err.Code(), err.Error())
		return
	}
	if len(comment.Comments) == 0 {
		c.AbortWithStatus(204)
		return
	}
	c.JSON(200, comment)
}

func addEntrepriseComment(c *gin.Context) {
	var comment Comment
	if c.ShouldBind(&comment) != nil {
		c.JSON(400, "malformed query")
		return
	}
	siret := c.Param("siret")
	username := c.GetString("username")
	comment.Siret = &siret
	comment.Username = &username
	err := comment.save()
	if err != nil {
		c.JSON(err.Code(), err.Error())
		return
	}
	c.JSON(200, comment)
}

func updateEntrepriseComment(c *gin.Context) {
	var message string
	id, err := strconv.Atoi(c.Param("id"))

	if c.ShouldBind(&message) != nil || err != nil {
		c.JSON(400, "malformed query")
		return
	}
	username := c.GetString("username")
	comment := Comment{
		ID:       &id,
		Username: &username,
		Message:  &message,
	}
	jerr := comment.update()
	if jerr != nil {
		c.JSON(jerr.Code(), jerr.Error())
		return
	}
	c.JSON(200, comment)

}

func (c *Comment) save() Jerror {
	sqlSaveComment := `insert into etablissement_comments
	(id_parent, siret, siren, username, message_history) 
	select $1, $2, substring($3 from 0 for 9), $5, array[$6]
	from etablissement0 e 
	left join etablissement_comments m on m.id = $7 and m.siret = e.siret
	where e.siret = $4 and (m.id is not null or $7 is null)
	returning id, siret, date_history, message_history;`

	err := db.QueryRow(
		context.Background(),
		sqlSaveComment,
		c.IDParent,
		c.Siret,
		c.Siret,
		c.Siret,
		c.Username,
		c.Message,
		c.IDParent,
	).Scan(&c.ID, &c.Siret, &c.DateHistory, &c.MessageHistory)
	c.Message = nil

	user, _ := getUser(*c.Username)
	if user.Username == nil {
		user.Username = c.Username
	}
	c.User = &user
	c.Username = nil

	if err != nil {
		if err.Error() == "no rows in result set" {
			return newJSONerror(403, "wrong SIRET")
		}
		return errorToJSON(500, err)
	}
	return nil
}

func (c *Comment) load() Jerror {
	sqlListComment := `select 
	e.id, id_parent, username, date_history, message_history, u.firstname, u.lastname
	from etablissement_comments e
	left join users u on u.username = e.username
	where e.siret = $1 order by e.id`

	rows, err := db.Query(context.Background(), sqlListComment, c.Siret)
	if err != nil {
		return errorToJSON(500, err)
	}

	comments := make(map[int]*Comment)
	var order []int

	for rows.Next() {
		var c Comment

		c.User = &user{}
		err := rows.Scan(&c.ID, &c.IDParent, &c.User.Username, &c.DateHistory, &c.MessageHistory, &c.User.FirstName, &c.User.LastName)
		if err != nil {
			return errorToJSON(500, err)
		}
		comments[*c.ID] = &c
		order = append(order, *c.ID)
	}

	for _, id := range order {
		v := comments[id]
		if v.IDParent == nil {
			c.Comments = append(c.Comments, v)
		} else {
			parent := comments[*v.IDParent]
			parent.Comments = append(parent.Comments, v)
		}
	}

	return nil
}

func (c *Comment) update() Jerror {
	sqlUpdateComment := `update etablissement_comments set 
	 message_history = array[$1] || message_history,
	 date_history = current_timestamp::timestamp || date_history
	 where username = $2 and id = $3 and message_history[1] != $4
	 returning id, id_parent, siret, username, date_history, message_history, null`

	err := db.QueryRow(context.Background(), sqlUpdateComment,
		c.Message,
		c.Username,
		c.ID,
		c.Message).Scan(
		&c.ID,
		&c.IDParent,
		&c.Siret,
		&c.Username,
		&c.DateHistory,
		&c.MessageHistory,
		&c.Message,
	)

	user, _ := getUser(*c.Username)
	if user.Username == nil {
		user.Username = c.Username
	}
	c.User = &user
	c.Username = nil

	if err != nil {
		if err.Error() == "no rows in result set" {
			return newJSONerror(403, "either duplicate comment, wrong username or wrong ID")
		}
		return errorToJSON(500, err)
	}
	return nil
}
