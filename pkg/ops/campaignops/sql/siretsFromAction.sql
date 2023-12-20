select ce.siret, cea.username, coalesce(last(cea.detail order by date_action), '')
from campaign_etablissement ce
inner join campaign_etablissement_action cea on cea.id_campaign_etablissement = ce.id
where id_campaign=$1
group by ce.id, ce.siret, cea.username
having last(cea.action order by date_action)=$2
