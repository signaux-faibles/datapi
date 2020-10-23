alter table score add detail jsonb;

create view v_entreprise_follow as 
	select siren, username 
	from etablissement_follow 
	where active
	group by siren, username;

create function get_summary (
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
		in exclure_suivi boolean,          -- $15
		in effectif_min int,               -- $16
		in effectif_max int,               -- $17
		in suivi_uniquement boolean,       -- $18
		in sirens text[]                   -- $19
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
		d.code as code_departement,
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
		inner join v_roles r on et.siren = r.siren
		inner join entreprise0 en on en.siren = r.siren
		inner join departements d on d.code = et.departement
		left join v_naf n on n.code_n5 = et.code_activite
		left join score s on et.siret = s.siret and	coalesce(s.libelle_liste, $4) = $4
		left join v_alert_etablissement aet on aet.siret = et.siret
		left join v_alert_entreprise aen on aen.siren = et.siren
		left join v_last_effectif ef on ef.siret = et.siret
		left join v_hausse_urssaf u on u.siret = et.siret
		left join v_apdemande ap on ap.siret = et.siret
		left join v_last_procol ep on ep.siret = et.siret
		left join v_diane_variation_ca di on di.siren = s.siren
		left join entreprise_ellisphere0 g on g.siren = et.siren
		left join terrind ti on ti.code_commune = et.code_commune
		left join v_entreprise_follow fe on fe.siren = et.siren and fe.username = $9
    left join etablissement_follow f on f.siret = et.siret and f.username = $9 and f.active
	where 
		(en.raison_sociale like $6 or et.siret like $5 or coalesce($5, $6) is null)
	  and (r.roles && $1 or $7)
    and (et.departement=any($1) or $8)
		and (et.siege or not $10)
		and (s.alert != 'Pas d''alerte' or not $12)
		and (coalesce(ep.last_procol, 'in_bonis') = any($13) or $13 is null)
		and (et.departement=any($14) or $14 is null)
		and (ef.effectif >= $16 or $16 is null)
		and (ef.effectif <= $17 or $17 is null)
		and (f.username = $9 and f.active or not $18)
		and (en.siren = any($19) or $19 is null)
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
$$ language sql immutable