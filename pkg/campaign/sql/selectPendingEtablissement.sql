with actions as (
  select id_campaign_etablissement,
    last(action order by id) as action
  from campaign_etablissement_action
  group by id_campaign_etablissement
)
select *
from campaign_etablissement ce
  inner join v_summaries s on s.siret = ce.siret and s.code_departement = any($2)
  left join actions a on a.id_campaign_etablissement = ce.id
where ce.id_campaign = $1 and coalesce(a.action, 'cancel') = 'cancel'
limit $3 offset $4;

