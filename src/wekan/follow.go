package wekan

import (
	"context"
	"github.com/signaux-faibles/datapi/src/db"
	"github.com/signaux-faibles/libwekan"
)

func followSiretsFromWekan(ctx context.Context, username libwekan.Username, sirets []string) error {
	tx, err := db.Get().Begin(ctx)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, sqlCreateTmpFollowWekan, sirets, username); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, sqlFollowFromTmp, username); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// sqlCreateTmpFollowWekan
const sqlCreateTmpFollowWekan = `create temporary table tmp_follow_wekan on commit drop as
	with sirets as (select unnest($1::text[]) as siret),
	follow as (select siret from etablissement_follow f where f.username = $2 and active)
	select case when f.siret is null then 'follow' else 'unfollow' end as todo,
	coalesce(f.siret, s.siret) as siret
	from follow f
	full join sirets s on s.siret = f.siret
	where f.siret is null or s.siret is null;
`

// sqlFollowFromTmp $1 = username
const sqlFollowFromTmp = `insert into etablissement_follow
	(siret, siren, username, active, since, comment, category)
	select t.siret, substring(t.siret from 1 for 9), $1,
	true, current_timestamp, 'participe à la carte wekan', 'wekan'
	from tmp_follow_wekan t
	inner join etablissement0 e on e.siret = t.siret
	where todo = 'follow'`
