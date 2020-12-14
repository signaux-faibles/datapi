drop materialized view v_summaries;
create materialized view v_summaries as
  with last_liste as (select first(libelle order by batch desc, algo desc) as last_liste from liste)
  select 
    r.roles, et.siret, et.siren, en.raison_sociale, et.commune,
    d.libelle as libelle_departement, et.departement as code_departement,
    s.score as valeur_score,
    s.detail as detail_score,
    aet.first_list = l.last_liste as first_alert,
    aen.first_list as first_list_entreprise,
  	l.last_liste as liste,
    di.chiffre_affaire, di.arrete_bilan, di.exercice_diane, di.variation_ca, di.resultat_expl, ef.effectif,
    n.libelle_n5, n.libelle_n1, et.code_activite, coalesce(ep.last_procol, 'in_bonis') as last_procol,
    coalesce(ap.ap, false) as activite_partielle,
    ap.heure_consomme as apconso_heure_consomme,
    ap.montant as apconso_montant,
    u.dette[1] > u.dette[2] or u.dette[2] > u.dette[3] as hausse_urssaf,
    u.dette[1] as dette_urssaf,
    s.alert,
    et.siege, g.raison_sociale as raison_sociale_groupe,
    ti.code_commune is not null as territoire_industrie
  from last_liste l
  	inner join etablissement0 et on true
    inner join entreprise0 en on en.siren = et.siren
    inner join v_etablissement_raison_sociale etrs on etrs.id_etablissement = et.id
    inner join v_roles r on et.siren = r.siren
    inner join departements d on d.code = et.departement
    left join v_naf n on n.code_n5 = et.code_activite
    left join score0 s on et.siret = s.siret and s.libelle_liste = l.last_liste
    left join v_alert_etablissement aet on aet.siret = et.siret
    left join v_alert_entreprise aen on aen.siren = et.siren
    left join v_last_effectif ef on ef.siret = et.siret
    left join v_hausse_urssaf u on u.siret = et.siret
    left join v_apdemande ap on ap.siret = et.siret
    left join v_last_procol ep on ep.siret = et.siret
    left join v_diane_variation_ca di on di.siren = et.siren
    left join entreprise_ellisphere0 g on g.siren = et.siren
    left join terrind ti on ti.code_commune = et.code_commune;

create or replace function get_score (
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
    exercice_diane int,
    variation_ca real,
    resultat_expl real,
    effectif real,
    libelle_n5 text,
    libelle_n1 text,
    code_activite text,
    last_procol text,
    activite_partielle boolean,
    apconso_heure_consomme int,
    apconso_montant int,
    hausse_urssaf boolean,
    dette_urssaf real,
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
  select 
    s.siret, s.siren, s.raison_sociale, s.commune,
    s.libelle_departement, s.code_departement,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.valeur_score end as valeur_score,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.detail_score end as detail_score,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.first_alert end as first_alert,
    s.chiffre_affaire, s.arrete_bilan, s.exercice_diane, s.variation_ca, s.resultat_expl, s.effectif,
    s.libelle_n5, s.libelle_n1, s.code_activite, s.last_procol,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp then s.activite_partielle end as activite_partielle,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp then s.apconso_heure_consomme end as apconso_heure_consomme,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp then s.apconso_montant end as apconso_montant,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).urssaf then s.hausse_urssaf end as hausse_urssaf,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).urssaf then s.dette_urssaf end as dette_urssaf,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.alert end,
    count(*) over () as nb_total,
    count(case when s.alert='Alerte seuil F1' and (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then 1 end) over () as nb_f1,
    count(case when s.alert='Alerte seuil F2' and (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then 1 end) over () as nb_f2,
    (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).visible, 
	(permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).in_zone, 
	  f.id is not null as followed_etablissement,
    fe.siren is not null as followed_entreprise,
    s.siege, s.raison_sociale_groupe, territoire_industrie,
    f.comment, f.category, f.since,
    (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).urssaf,
    (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp,
    (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score,
    (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).bdf
  from v_summaries s
    left join etablissement_follow f on f.active and f.siret = s.siret and f.username = $9
    left join v_entreprise_follow fe on fe.siren = s.siren and fe.username = $9
    where   
    (s.roles && $1 or $7)
	and (s.code_departement=any($1) or $8)
    and (s.siege or not $10)
    and (s.alert != 'Pas d''alerte')
    and (s.last_procol = any($13) or $13 is null)
    and (s.code_departement=any($14) or $14 is null)
    and (s.effectif >= $16 or $16 is null)
    and (s.effectif <= $17 or $17 is null)
	and (s.raison_sociale ilike $6 or s.siret ilike $5 or coalesce($5, $6) is null)
    and 'score' = any($1)
  order by s.valeur_score desc, s.siret
  limit $2 offset $3
$$ language sql immutable;
