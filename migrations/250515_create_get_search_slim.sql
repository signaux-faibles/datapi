DROP FUNCTION IF EXISTS public.get_search_slim(
    text[], int4, int4, text, text, text, bool, bool, text, bool, text, bool,
    text[], text[], bool, int4, int4, text[], text[], int4, int4, int4, int4, text[], text
);

CREATE OR REPLACE FUNCTION public.get_search_slim(
    roles_users text[], nblimit integer, nboffset integer, libelle_liste text,
    siret_expression text, raison_sociale_expression text, ignore_roles boolean,
    ignore_zone boolean, username text, siege_uniquement boolean, order_by text,
    alert_only boolean, last_procol text[], departements text[], suivi boolean,
    effectif_min integer, effectif_max integer, sirens text[], activites text[],
    effectif_min_entreprise integer, effectif_max_entreprise integer, ca_min integer,
    ca_max integer, exclude_secteurs_covid text[], etat_administratif text
)
RETURNS TABLE(
    siret text, siren text, raison_sociale text, commune text, libelle_departement text,
    code_departement text, valeur_score real, detail_score jsonb, first_alert boolean,
    chiffre_affaire real, arrete_bilan date, exercice_diane integer, variation_ca real,
    resultat_expl real, effectif real, effectif_entreprise real, libelle_n5 text,
    libelle_n1 text, code_activite text, last_procol text, activite_partielle boolean,
    apconso_heure_consomme integer, apconso_montant integer, hausse_urssaf boolean,
    dette_urssaf real, alert text, nb_total bigint, nb_f1 bigint, nb_f2 bigint,
    visible boolean, in_zone boolean, followed boolean, followed_enterprise boolean,
    siege boolean, raison_sociale_groupe text, territoire_industrie boolean,
    comment text, category text, since timestamp without time zone, urssaf boolean,
    dgefp boolean, score boolean, bdf boolean, secteur_covid text,
    excedent_brut_d_exploitation real, etat_administratif text,
    etat_administratif_entreprise text
)
LANGUAGE sql
IMMUTABLE
AS $function$
SELECT
    s.siret, s.siren, s.raison_sociale, s.commune,
    s.libelle_departement, s.code_departement,
    CASE WHEN (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren IS NOT NULL)).score THEN s.valeur_score END AS valeur_score,
    CASE WHEN (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren IS NOT NULL)).score THEN s.detail_score END AS detail_score,
    CASE WHEN (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren IS NOT NULL)).score THEN s.first_alert END AS first_alert,
    s.chiffre_affaire, s.arrete_bilan, s.exercice_diane, s.variation_ca, s.resultat_expl, s.effectif, s.effectif_entreprise,
    s.libelle_n5, s.libelle_n1, s.code_activite, s.last_procol,
    CASE WHEN (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren IS NOT NULL)).dgefp THEN s.activite_partielle END AS activite_partielle,
    CASE WHEN (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren IS NOT NULL)).dgefp THEN s.apconso_heure_consomme END AS apconso_heure_consomme,
    CASE WHEN (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren IS NOT NULL)).dgefp THEN s.apconso_montant END AS apconso_montant,
    CASE WHEN (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren IS NOT NULL)).urssaf THEN s.hausse_urssaf END AS hausse_urssaf,
    CASE WHEN (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren IS NOT NULL)).urssaf THEN s.dette_urssaf END AS dette_urssaf,
    CASE WHEN (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren IS NOT NULL)).score THEN s.alert END AS alert,
    NULL::bigint AS nb_total,
    NULL::bigint AS nb_f1,
    NULL::bigint AS nb_f2,
    (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren IS NOT NULL)).visible,
    (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren IS NOT NULL)).in_zone,
    f.id IS NOT NULL AS followed_etablissement,
    fe.siren IS NOT NULL AS followed_entreprise,
    s.siege, s.raison_sociale_groupe, territoire_industrie,
    f.comment, f.category, f.since,
    (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren IS NOT NULL)).urssaf,
    (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren IS NOT NULL)).dgefp,
    (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren IS NOT NULL)).score,
    (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren IS NOT NULL)).bdf,
    s.secteur_covid, s.excedent_brut_d_exploitation, s.etat_administratif, s.etat_administratif_entreprise
FROM v_summaries s
    LEFT JOIN etablissement_follow f ON f.active AND f.siret = s.siret AND f.username = $9
    LEFT JOIN v_entreprise_follow fe ON fe.siren = s.siren AND fe.username = $9
    LEFT JOIN v_naf n ON n.code_n5 = s.code_activite
WHERE
    (s.raison_sociale ILIKE $6 OR s.siret ILIKE $5)
    AND (s.roles && $1 OR $7)
    AND (s.code_departement = ANY($1) OR $8)
    AND (s.code_departement = ANY($14) OR $14 IS NULL)
    AND (s.effectif >= $16 OR $16 IS NULL)
    AND (s.effectif_entreprise >= $20 OR $20 IS NULL)
    AND (s.effectif_entreprise <= $21 OR $21 IS NULL)
    AND (s.chiffre_affaire >= $22 OR $22 IS NULL)
    AND (s.chiffre_affaire <= $23 OR $23 IS NULL)
    AND (n.code_n1 = ANY($19) OR $19 IS NULL)
    AND (s.siege OR NOT $10)
    AND (NOT (s.secteur_covid = ANY($24)) OR $24 IS NULL)
    AND (s.etat_administratif = $25 OR $25 IS NULL)
  order by s.raison_sociale, s.siret
LIMIT $2 OFFSET $3;
$function$;
