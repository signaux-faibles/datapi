insert into campaign_etablissement_action
  (id_campaign_etablissement, username, date_action, action, detail)
with actions as (
  select id_campaign_etablissement,
         last(action order by id) as action,
         last(username order by id) as username
  from campaign_etablissement_action
  group by id_campaign_etablissement
)
select ce.id, $3, current_timestamp, $5, $6
from campaign_etablissement ce
       inner join v_summaries s on s.siret = ce.siret and code_departement = any($4)
       left join actions a on a.id_campaign_etablissement = ce.id
where ce.id_campaign = $1 and ce.id = $2 and a.action = 'take' and a.username = $3
returning id;
