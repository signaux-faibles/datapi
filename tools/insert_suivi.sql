-- insert des siren à suivre
drop table if exists insert_entreprise_suivi;
create temporary table insert_entreprise_suivi (siren text);
insert into insert_entreprise_suivi values
('123456789'),
('234567890');

-- insert des sirets à suivre
drop table if exists insert_etablissement_suivi;
create temporary table insert_etablissement_suivi (siret text);
insert into insert_etablissement_suivi values
('12345678901234'),
('23456789012345');

-- détection des établissements/entreprises absents
select 'etablissement manquant' as motif, i.siret from insert_etablissement_suivi i
left join etablissement0 e on e.siret = i.siret
where e.id is null
union
select 'entreprise manquante', i.siren from insert_entreprise_suivi i
left join entreprise0 e on e.siren = i.siren
where e.id is null

-- établissement sélectionné pour les entreprises, détection éventuelle des sieges manquants
select i.siren, 
coalesce(case when en.siren is null then 'entreprise manquante' end,
		 case when e.siret is null then 'pas de siege pour cette entreprise' end,
		 e.siret)
from insert_entreprise_suivi i
left join entreprise0 en on en.siren = i.siren
left join etablissement0 e on e.siren = i.siren and e.siege


-- ajout du suivi des établissements qui ne sont pas encore (ou plus) suivis pour emmanuel.lemaux
with to_follow as (
select s.siret from insert_etablissement_suivi s
inner join etablissement e on e.siret = s.siret
union select e.siret from insert_entreprise_suivi s
inner join etablissement e on e.siren = s.siren and e.siege)
insert into etablissement_follow (siret, siren, username, active, since, comment, category)
select t.siret, substring(t.siret from 1 for 9),
'email@user', true, current_timestamp, 'Suivi extrait du fichier de suivi CRP', 'fichier_CRP'
from to_follow t
left join etablissement_follow f on f.siret = t.siret and f.username = 'email@user' and until is null and active
where f.id is null
