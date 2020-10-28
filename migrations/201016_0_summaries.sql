drop materialized view v_alert_etablissement;
drop materialized view v_alert_entreprise;

alter table score add detail jsonb;

drop view score0;
create view score0 as
select * from score where version = 0;

create materialized view v_alert_etablissement as
select siret, min(libelle_liste) as first_list
from score0
where alert in ('Alerte seuil F1', 'Alerte seuil F2')
group by siret;
create unique index idx_v_alert_etablissement_siret
  on v_alert_etablissement (siret);

create materialized view v_alert_entreprise as
select siren, min(libelle_liste) as first_list
from score0
where alert in ('Alerte seuil F1', 'Alerte seuil F2')
group by siren;
create index idx_v_alert_entreprise_siren
  on v_alert_entreprise (siren);

create materialized view v_etablissement_raison_sociale as 
select et.id as id_etablissement, en.id as id_entreprise, en.raison_sociale, et.siret
from entreprise0 en
inner join etablissement0 et on en.siren=et.siren;

create extension pg_trgm;
create index idx_entreprise_raison_sociale on v_etablissement_raison_sociale using gin (siret gin_trgm_ops, raison_sociale gin_trgm_ops);
create index idx_entreprise_ellisphere_siren on entreprise_ellisphere (siren) where version = 0;
create index idx_entreprise_diane_siren on entreprise_diane (siren) where version = 0;
create index idx_entreprise_bdf_siren on entreprise_bdf (siren) where version = 0;
create index idx_etablissement_follow_siret_username_active on etablissement_follow (siret, siren, username, active);
create index idx_etablissement_follow_siren_username on etablissement_follow (active, siren, username);
create index idx_score_siret_liste on score (siret, libelle_liste, alert) where version = 0;

create view v_entreprise_follow as 
	select siren, username 
	from etablissement_follow 
	where active
	group by siren, username;

create or replace function f_etablissement_permissions (
	in roles_user text[], 
	in username text
) returns table (
	id int,
	siren varchar(9),
	siret varchar(14),
	followed boolean,
	visible boolean,
	in_zone boolean,
	score boolean,
	urssaf boolean,
	dgefp boolean,
	bdf boolean
) as $$
select e.id, e.siren, e.siret, f.siren is not null as followed,
(permissions($1, r.roles, a.first_list, e.departement, f.siren is not null)).*
from etablissement0 e
inner join v_roles r on r.siren = e.siren
left join v_entreprise_follow f on f.siren = e.siren and f.username = $2
left join v_alert_entreprise a on a.siren = e.siren
$$ language sql immutable;

create or replace function get_summary (
		in roles_users text[],             -- $1
		in nblimit int,                    -- $2
		in nboffset int,                   -- $3
		in libelle_liste text,             -- $4
		in siret_expression text,          -- $5
		in raison_sociale_expression text, -- $6
		in ignore_roles boolean,           -- $7
		in ignore_zone boolean,            -- $8
		in username text,                  -- $9
		in siege_uniquement boolean,       -- $10
		in order_by text,                  -- $11
    in alert_only boolean,             -- $12
		in last_procol text[],             -- $13
		in departements text[],            -- $14
		in suivi boolean,                  -- $15
		in effectif_min int,               -- $16
		in effectif_max int,               -- $17
		in sirens text[]                   -- $18
	) returns table (
		siret text,
		siren text,
		raison_sociale text,
		commune text,
		libelle_departement text,
		code_departement text,
		valeur_score real,
		detail_score jsonb,
		first_alert boolean,
		chiffre_affaire real,
		arrete_bilan date,
		variation_ca real,
		resultat_expl real,
		effectif real,
		libelle_n5 text,
		libelle_n1 text,
		code_activite text,
		last_procol text,
		activite_partielle boolean,
		hausse_urssaf boolean,
		alert text,
		nb_total bigint,
		nb_f1 bigint,
		nb_f2 bigint,
		visible boolean,
		in_zone boolean,
		followed boolean,
    followed_enterprise boolean,
		siege boolean,
		raison_sociale_groupe text,
		territoire_industrie boolean,
		comment text,
		category text,
		since timestamp,
    urssaf boolean,
    dgefp boolean,
    score boolean,
		bdf boolean
) as $$
with open_summary as (
	select 
		et.siret,
		et.siren,
		en.raison_sociale,
		et.commune,
		d.libelle as libelle_departement,
		et.departement as code_departement,
		s.score as valeur_score,
		s.detail as detail_score,
		aet.first_list = $4 as first_alert,
		di.chiffre_affaire,
		di.arrete_bilan,
		di.variation_ca,
		di.resultat_expl,
		ef.effectif,
		n.libelle_n5,
		n.libelle_n1,
		et.code_activite,
		coalesce(ep.last_procol, 'in_bonis') as last_procol,
		coalesce(ap.ap, false) activite_partielle,
		u.dette[0] > u.dette[1] or u.dette[1] > u.dette[2] as hausse_urssaf,
		s.alert,
		f.id is not null as followed_etablissement,
    fe.siren is not null as followed_entreprise,
		count(case when s.alert='Alerte seuil F1' then 1 else null end) over () as nb_f1,
		count(case when s.alert='Alerte seuil F2' then 1 else null end) over () as nb_f2,
		count(*) over () as nb_total,
		et.siege,
		g.raison_sociale as raison_sociale_groupe,
		ti.code_commune is not null as territoire_industrie,
		f.comment, f.category, f.since,
		(permissions($1, r.roles, aen.first_list, d.code, fe.siren is not null)).*
	from etablissement0 et
		inner join entreprise0 en on en.siren = et.siren
		inner join v_etablissement_raison_sociale etrs on etrs.id_etablissement = et.id
		inner join v_roles r on et.siren = r.siren
		inner join departements d on d.code = et.departement
    left join etablissement_follow f on f.active and f.siret = et.siret and f.username = $9
		left join v_entreprise_follow fe on fe.siren = et.siren and fe.username = $9
		left join v_naf n on n.code_n5 = et.code_activite
		left join score0 s on et.siret = s.siret and s.libelle_liste = $4
		left join v_alert_etablissement aet on aet.siret = et.siret
		left join v_alert_entreprise aen on aen.siren = et.siren
		left join v_last_effectif ef on ef.siret = et.siret
		left join v_hausse_urssaf u on u.siret = et.siret
		left join v_apdemande ap on ap.siret = et.siret
		left join v_last_procol ep on ep.siret = et.siret
		left join v_diane_variation_ca di on di.siren = et.siren
		left join entreprise_ellisphere0 g on g.siren = et.siren
		left join terrind ti on ti.code_commune = et.code_commune
	where 
		(etrs.raison_sociale ilike $6 or etrs.siret ilike $5 or coalesce($5, $6) is null)
	  and (r.roles && $1 or $7)
    and (et.departement=any($1) or $8)
		and (et.siege or not $10)
		and (s.alert != 'Pas d''alerte' or not $12)
		and (coalesce(ep.last_procol, 'in_bonis') = any($13) or $13 is null)
		and (et.departement=any($14) or $14 is null)
		and (ef.effectif >= $16 or $16 is null)
		and (ef.effectif <= $17 or $17 is null)
		and (f.username is not null = $15 or $15 is null)
		and (en.siren = any($18) or $18 is null)
	order by case when $11 = 'score' then s.score end desc,
	         case when $11 = 'raison_sociale' then en.raison_sociale end,
		 	 case when $11 = 'follow' then f.id end,
			 case when $11 = 'effectif_desc' then ef.effectif end desc, 
			 et.siret
	limit $2 offset $3
) select siret, 
		siren, 
		raison_sociale, 
		commune, 
		libelle_departement, 
		code_departement, 
		valeur_score,
		detail_score,
		first_alert,
		chiffre_affaire,
		arrete_bilan,
		variation_ca,
		resultat_expl,
		effectif,
		libelle_n5,
		libelle_n1,
		code_activite,
		last_procol,
		case when dgefp then activite_partielle else null end as activite_partielle,
		case when urssaf then hausse_urssaf else null end as hausse_urssaf,
		case when score then alert else null end,
		nb_total,
		nb_f1,
		nb_f2,
		visible, 
		in_zone, 
		followed_etablissement, 
		followed_entreprise,
		siege, 
		raison_sociale_groupe, 
		territoire_industrie,
		comment,
		category,
		since,
		urssaf,
		dgefp,
		score,
		bdf from open_summary
$$ language sql immutable;