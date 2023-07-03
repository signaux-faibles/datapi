with actions as (
  select id_campaign_etablissement,
    last(username order by id) as username,
    last(action order by id) as action
  from campaign_etablissement_action
  group by id_campaign_etablissement
)
select count(*) over () as nb_total,
   s.siret,
   s.raison_sociale,
   s.raison_sociale_groupe,
   s.alert,
   ce.id,
   ce.id_campaign,
   f.id is not null as followed,
   s.first_alert,
   s.etat_administratif,
   a.action,
   rank() over (order by ce.id) as rank,
   a.username
from campaign_etablissement ce
  inner join v_summaries s on s.siret = ce.siret and s.code_departement = any($2)
  left join etablissement_follow f on f.siret = ce.siret and f.username=$3 and f.active
  left join actions a on a.id_campaign_etablissement = ce.id
where ce.id_campaign = $1


