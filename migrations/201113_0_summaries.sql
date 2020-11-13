drop index idx_score_score;
create index idx_score_score on score (score desc) where version=0;

create function get_score (
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
	select 
    et.siret,
    et.siren,
    en.raison_sociale,
    et.commune,
    d.libelle as libelle_departement,
    et.departement as code_departement,
    case when p.score then s.score end as valeur_score,
	  case when p.score then s.detail end as detail_score,
    case when p.score then aet.first_list = $4 end as first_alert,
    di.chiffre_affaire,
    di.arrete_bilan,
    di.variation_ca,
    di.resultat_expl,
    ef.effectif,
    n.libelle_n5,
    n.libelle_n1,
    et.code_activite,
    coalesce(ep.last_procol, 'in_bonis') as last_procol,
    case when p.dgefp then coalesce(ap.ap, false) end activite_partielle,
    case when p.urssaf then u.dette[0] > u.dette[1] or u.dette[1] > u.dette[2] end as hausse_urssaf,
    case when p.score then s.alert end,
    count(*) over () as nb_total,
    count(case when s.alert='Alerte seuil F1' then 1 else null end) over () as nb_f1,
    count(case when s.alert='Alerte seuil F2' then 1 else null end) over () as nb_f2,
    p.visible,
    p.in_zone,
    f.id is not null as followed,
    fe.siren is not null as followed_entreprise,
    et.siege,
    g.raison_sociale as raison_sociale_groupe,
    ti.code_commune is not null as territoire_industrie,
    f.comment, f.category, f.since,
    p.urssaf, p.dgefp, p.score, p.bdf
	from etablissement0 et
    inner join entreprise0 en on en.siren = et.siren
    inner join f_etablissement_permissions($1, $9) p on et.id = p.id
    inner join v_etablissement_raison_sociale etrs on etrs.id_etablissement = et.id
    inner join score0 s on et.siret = s.siret and s.libelle_liste = $4
    inner join departements d on d.code = et.departement
    left join etablissement_follow f on f.active and f.siret = et.siret and f.username = $9
    left join v_entreprise_follow fe on fe.siren = et.siren and fe.username = $9
    left join v_naf n on n.code_n5 = et.code_activite
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
	  and p.visible
    and (p.in_zone or $8)
    and (et.siege or not $10)
    and (s.alert != 'Pas d''alerte' or not $12)
    and (coalesce(ep.last_procol, 'in_bonis') = any($13) or $13 is null)
    and (et.departement=any($14) or $14 is null)
    and (ef.effectif >= $16 or $16 is null)
    and (ef.effectif <= $17 or $17 is null)
    and (f.username is not null = $15 or $15 is null)
    and (en.siren = any($18) or $18 is null)
	order by s.score desc, et.siret
	limit $2 offset $3
$$ language sql immutable;