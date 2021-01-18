DROP INDEX public.idx_score_liste;

CREATE INDEX idx_score_liste
    ON public.score USING btree
    (siret COLLATE pg_catalog."default" ASC NULLS LAST, score ASC NULLS LAST, libelle_liste COLLATE pg_catalog."default" ASC NULLS LAST)
    TABLESPACE pg_default
    WHERE version = 0 AND alert <> 'pas d''alerte'::text;
-- FUNCTION: public.get_score(text[], integer, integer, text, text, text, boolean, boolean, text, boolean, text, boolean, text[], text[], boolean, integer, integer, text[], text[])

-- DROP FUNCTION public.get_score(text[], integer, integer, text, text, text, boolean, boolean, text, boolean, text, boolean, text[], text[], boolean, integer, integer, text[], text[]);

CREATE OR REPLACE FUNCTION public.get_score(
	roles_users text[],
	nblimit integer,
	nboffset integer,
	libelle_liste text,
	siret_expression text,
	raison_sociale_expression text,
	ignore_roles boolean,
	ignore_zone boolean,
	username text,
	siege_uniquement boolean,
	order_by text,
	alert_only boolean,
	last_procol text[],
	departements text[],
	suivi boolean,
	effectif_min integer,
	effectif_max integer,
	sirens text[],
	activites text[])
    RETURNS TABLE(siret text, siren text, raison_sociale text, commune text, libelle_departement text, code_departement text, valeur_score real, detail_score jsonb, first_alert boolean, chiffre_affaire real, arrete_bilan date, exercice_diane integer, variation_ca real, resultat_expl real, effectif real, libelle_n5 text, libelle_n1 text, code_activite text, last_procol text, activite_partielle boolean, apconso_heure_consomme integer, apconso_montant integer, hausse_urssaf boolean, dette_urssaf real, alert text, nb_total bigint, nb_f1 bigint, nb_f2 bigint, visible boolean, in_zone boolean, followed boolean, followed_enterprise boolean, siege boolean, raison_sociale_groupe text, territoire_industrie boolean, comment text, category text, since timestamp without time zone, urssaf boolean, dgefp boolean, score boolean, bdf boolean) 
    LANGUAGE 'sql'
    COST 100
    IMMUTABLE PARALLEL UNSAFE
    ROWS 1000

AS $BODY$
select 
    s.siret, s.siren, s.raison_sociale, s.commune,
    s.libelle_departement, s.code_departement,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then sc.score end as valeur_score,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then sc.detail end as detail_score,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.first_alert end as first_alert,
    s.chiffre_affaire, s.arrete_bilan, s.exercice_diane, s.variation_ca, s.resultat_expl, s.effectif,
    s.libelle_n5, s.libelle_n1, s.code_activite, s.last_procol,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp then s.activite_partielle end as activite_partielle,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp then s.apconso_heure_consomme end as apconso_heure_consomme,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp then s.apconso_montant end as apconso_montant,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).urssaf then s.hausse_urssaf end as hausse_urssaf,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).urssaf then s.dette_urssaf end as dette_urssaf,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then sc.alert end,
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
    inner join score0 sc on sc.siret = s.siret and sc.libelle_liste = $4 and sc.alert != 'Pas d''alerte'
    left join v_naf n on n.code_n5 = s.code_activite
    left join etablissement_follow f on f.active and f.siret = s.siret and f.username = $9
    left join v_entreprise_follow fe on fe.siren = s.siren and fe.username = $9
    where   
    (s.roles && $1 or $7)
	and (s.code_departement=any($1) or $8)
    and (s.siege or not $10)
    and (s.last_procol = any($13) or $13 is null)
    and (s.code_departement=any($14) or $14 is null)
    and (s.effectif >= $16 or $16 is null)
    and (s.effectif <= $17 or $17 is null)
    and (fe.siren is not null = $15 or $15 is null)
	  and (s.raison_sociale ilike $6 or s.siret ilike $5 or coalesce($5, $6) is null)
    and 'score' = any($1)
    and (n.code_n1 = any($19) or $19 is null)
  order by sc.score desc, sc.siret
  limit $2 offset $3
$BODY$;

ALTER FUNCTION public.get_score(text[], integer, integer, text, text, text, boolean, boolean, text, boolean, text, boolean, text[], text[], boolean, integer, integer, text[], text[])
    OWNER TO datapi;
