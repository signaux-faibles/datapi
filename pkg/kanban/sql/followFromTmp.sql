insert into etablissement_follow
(siret, siren, username, active, since, comment, category)
select t.siret, substring(t.siret from 1 for 9), $1,
       true, current_timestamp, 'participe à la carte wekan', 'wekan'
from tmp_follow_wekan t
       inner join etablissement0 e on e.siret = t.siret
where todo = 'follow';
