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
    in alert_only boolean              -- $12
	) returns table (
		siret text,
		siren text,
		raison_sociale text,
		commune text,
		libelle_departement text,
		code_departement text,
		score real,
		detail_score jsonb,
		chiffre_affaire real,
		arrete_bilan date,
		variation_ca real,
		resultat_expl real,
		effectif real,
		libelle_n5 text,
		libelle_n1 text,
		code_activite text,
		last_procol text,
		activite_partielle bool,
		hausse_urssaf bool,
		alert text,
		nb_total bigint,
		visible bool,
		in_zone bool,
		followed bool,
		siege bool,
		raison_sociale_groupe text,
		territoire_industrie bool
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
		f.siren is not null as followed,
		count(*) over () as nb_total,
		et.siege,
		g.raison_sociale as raison_sociale_groupe,
		ti.code_commune is not null as territoire_industrie,
		(permissions($1, r.roles, aen.first_list, d.code, f.siren is not null)).*
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
		left join v_entreprise_follow f on f.siren = s.siren and f.username = $9
	where 
		(en.raison_sociale like $6 or et.siret like $5)
	    and (r.roles && $1 or $7)
        and (et.departement=any($1) or $8)
		and (et.siege or $10)
	order by case when $11 = 'score' then s.score end desc,
	         case when $11 = 'raison_sociale' then en.raison_sociale end, 
			 case when $11 = 'raison_sociale' then et.siret end
	limit $2 offset $3
) select siret, 
	siren, 
	raison_sociale, 
	commune, 
	libelle_departement, 
	code_departement, 
	valeur_score, 
	detail_score,
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
	visible, 
	in_zone, 
	followed, 
	siege, 
	raison_sociale_groupe, territoire_industrie from open_summary
$$ language sql immutable


 