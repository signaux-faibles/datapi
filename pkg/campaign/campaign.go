package campaign

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

type CampaignID int

type Campaign struct {
	ID                CampaignID `json:"id"`
	Libelle           string     `json:"libelle"`
	DateEnd           time.Time  `json:"date_fin"`
	DateCreate        time.Time  `json:"date_create"`
	WekanDomainRegexp string     `json:"wekan_domain_regexp"`
	BoardSlugs        []string   `json:"slugs"`
	NBTotal           int        `json:"nb_total"`
	NBPerimetre       int        `json:"nb_perimetre"`
}

type Campaigns []*Campaign

func (cs *Campaigns) Tuple() []interface{} {
	c := Campaign{}
	*cs = append(*cs, &c)
	return []interface{}{
		&c.ID,
		&c.Libelle,
		&c.WekanDomainRegexp,
		&c.DateEnd,
		&c.DateCreate,
		&c.BoardSlugs,
		&c.NBTotal,
		&c.NBPerimetre,
	}
}

func (cs *Campaign) Insert(ctx context.Context, db *pgxpool.Pool) (CampaignID, error) {
	sql := `insert into campaign (libelle, wekan_domain_regexp, date_end, date_create)
          values ($1, $2, $3, $4) returning id`
	err := db.QueryRow(ctx, sql, cs.Libelle, cs.WekanDomainRegexp, cs.DateEnd, cs.DateCreate).Scan(&cs.ID)
	return cs.ID, err
}
