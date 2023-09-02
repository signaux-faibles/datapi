alter table entreprise_diane
drop column if exists version cascade,
drop column if exists date_add cascade,
drop column if exists nombre_etab_secondairecascade,
drop column if exists nombre_filialecascade,
drop column if exists nombre_moiscascade,
drop column if exists credit_client cascade,
drop column if exists achat_marchandises cascade,
drop column if exists achat_matieres_premieres  cascade,
drop column if exists autonomie_financiere cascade,
drop column if exists autres_achats_charges_externes cascade,
drop column if exists autres_produits_charges_reprises cascade,
drop column if exists ca_exportation cascade,
drop column if exists capacite_autofinancement cascade,
drop column if exists capacite_remboursement cascade,
drop column if exists charge_exceptionnelle cascade,
drop column if exists charge_personnel cascade,
drop column if exists charges_financieres cascade,
drop column if exists conces_brev_et_droits_sim cascade,
drop column if exists consommation cascade,
drop column if exists couverture_ca_besoin_fdr cascade,
drop column if exists couverture_ca_fdr cascade,
drop column if exists credit_fournisseur cascade,
drop column if exists degre_immo_corporelle cascade,
drop column if exists dette_fiscale_et_sociale cascade,
drop column if exists dotation_amortissement cascade,
drop column if exists endettement cascade,
drop column if exists endettement_global cascade,
drop column if exists equilibre_financier cascade,
drop column if exists exportation cascade,
drop column if exists financement_actif_circulant cascade,
drop column if exists frais_de_RetD cascade,
drop column if exists impot_benefice cascade,
drop column if exists impots_taxes cascade,
drop column if exists independance_financiere cascade,
drop column if exists interets cascade,
drop column if exists liquidite_generale cascade,
drop column if exists liquidite_reduite cascade,
drop column if exists marge_commerciale cascade,
drop column if exists operations_commun cascade,
drop column if exists part_autofinancement cascade,
drop column if exists part_etat cascade,
drop column if exists part_preteur cascade,
drop column if exists part_salaries cascade,
drop column if exists participation_salaries cascade,
drop column if exists performance cascade,
drop column if exists poids_bfr_exploitation cascade,
drop column if exists procedure_collective cascade,
drop column if exists production cascade,
drop column if exists productivite_capital_financier cascade,
drop column if exists productivite_capital_investi cascade,
drop column if exists productivite_potentiel_production cascade,
drop column if exists produit_exceptionnel cascade,
drop column if exists produits_financiers cascade,
drop column if exists rendement_brut_fonds_propres cascade,
drop column if exists rendement_capitaux_propres cascade,
drop column if exists rendement_ressources_durables cascade,
drop column if exists rentabilite_economique cascade,
drop column if exists rentabilite_nette cascade,
drop column if exists resultat_avant_impot cascade,
drop column if exists rotation_stocks cascade,
drop column if exists statut_juridique cascade,
drop column if exists subventions_d_exploitation cascade,
drop column if exists taille_compo_groupe cascade,
drop column if exists taux_d_investissement_productif cascade,
drop column if exists taux_endettement cascade,
drop column if exists taux_interet_financier cascade,
drop column if exists taux_interet_sur_ca cascade,
drop column if exists taux_valeur_ajoutee cascade,
drop column if exists valeur_ajoutee cascade,
drop column if exists hash cascade;

drop view if exists entreprise_diane0 cascade;
create or replace view entreprise_diane0
    (id, siren, arrete_bilan_diane, exercice_diane, chiffre_affaire, resultat_expl,
      excedent_brut_d_exploitation, benefice_ou_perte)
as SELECT id, siren, arrete_bilan_diane, exercice_diane, chiffre_affaire, resultat_expl,
      excedent_brut_d_exploitation, benefice_ou_perte
FROM entreprise_diane;

create materialized view v_diane_variation_ca as
WITH diane AS (SELECT entreprise_diane0.siren,
                      entreprise_diane0.arrete_bilan_diane,
                      entreprise_diane0.exercice_diane::integer                                                 AS exercice_diane,
                      lag(entreprise_diane0.arrete_bilan_diane, 1)
                      OVER (PARTITION BY entreprise_diane0.siren ORDER BY entreprise_diane0.arrete_bilan_diane) AS prev_arrete_bilan_diane,
                      entreprise_diane0.chiffre_affaire,
                      lag(entreprise_diane0.chiffre_affaire, 1)
                      OVER (PARTITION BY entreprise_diane0.siren ORDER BY entreprise_diane0.arrete_bilan_diane) AS prev_chiffre_affaire,
                      entreprise_diane0.resultat_expl,
                      lag(entreprise_diane0.resultat_expl, 1)
                      OVER (PARTITION BY entreprise_diane0.siren ORDER BY entreprise_diane0.arrete_bilan_diane) AS prev_resultat_expl,
                      entreprise_diane0.excedent_brut_d_exploitation,
                      lag(entreprise_diane0.excedent_brut_d_exploitation, 1)
                      OVER (PARTITION BY entreprise_diane0.siren ORDER BY entreprise_diane0.arrete_bilan_diane) AS prev_excedent_brut_d_exploitation
               FROM entreprise_diane0
               ORDER BY entreprise_diane0.siren, entreprise_diane0.arrete_bilan_diane DESC)
SELECT diane.siren,
       first(diane.arrete_bilan_diane ORDER BY diane.arrete_bilan_diane DESC)                AS arrete_bilan,
       first(diane.exercice_diane ORDER BY diane.arrete_bilan_diane DESC)                    AS exercice_diane,
       first(diane.prev_chiffre_affaire ORDER BY diane.arrete_bilan_diane DESC)              AS prev_chiffre_affaire,
       first(diane.chiffre_affaire ORDER BY diane.arrete_bilan_diane DESC)                   AS chiffre_affaire,
       first(
         CASE
           WHEN diane.prev_chiffre_affaire <> 0::double precision
             THEN diane.chiffre_affaire / diane.prev_chiffre_affaire
           ELSE NULL::real
           END ORDER BY diane.arrete_bilan_diane DESC)                               AS variation_ca,
       first(diane.resultat_expl ORDER BY diane.arrete_bilan_diane DESC)                     AS resultat_expl,
       first(diane.prev_resultat_expl ORDER BY diane.arrete_bilan_diane DESC)                AS prev_resultat_expl,
       first(diane.excedent_brut_d_exploitation
             ORDER BY diane.arrete_bilan_diane DESC)                                         AS excedent_brut_d_exploitation,
       first(diane.prev_excedent_brut_d_exploitation
             ORDER BY diane.arrete_bilan_diane DESC)                                         AS prev_excedent_brut_d_exploitation
FROM diane
GROUP BY diane.siren;

alter materialized view v_diane_variation_ca owner to postgres;

create unique index idx_v_diane_variation_ca_siren
  on v_diane_variation_ca (siren);

create materialized view v_summaries as
WITH last_liste AS (SELECT first(liste.libelle ORDER BY liste.batch DESC, liste.algo DESC) AS last_liste
                    FROM liste)
SELECT r.roles,
       et.siret,
       et.siren,
       en.raison_sociale,
       et.commune,
       d.libelle                                                                                           AS libelle_departement,
       et.departement                                                                                      AS code_departement,
       s.score                                                                                             AS valeur_score,
       s.detail                                                                                            AS detail_score,
       COALESCE(aet.first_list = l.last_liste OR aet.first_red_list = l.last_liste,
                false)                                                                                     AS first_alert,
       aen.first_list                                                                                      AS first_list_entreprise,
       aen.first_red_list                                                                                  AS first_red_list_entreprise,
       aet.first_list                                                                                      AS first_list_etablissement,
       aet.first_red_list                                                                                  AS first_red_list_etablissement,
       aet.last_list,
       aet.last_alert,
       l.last_liste                                                                                        AS liste,
       di.chiffre_affaire,
       di.prev_chiffre_affaire,
       di.arrete_bilan,
       di.exercice_diane,
       di.variation_ca,
       di.resultat_expl,
       di.prev_resultat_expl,
       di.excedent_brut_d_exploitation,
       di.prev_excedent_brut_d_exploitation,
       ef.effectif,
       sum(ef.effectif)
       FILTER (WHERE COALESCE(et.etat_administratif, 'A'::text) = 'A'::text) OVER (PARTITION BY en.siren)  AS effectif_entreprise,
       max(ef.periode)
       FILTER (WHERE COALESCE(et.etat_administratif, 'A'::text) <> 'A'::text) OVER (PARTITION BY en.siren) AS date_entreprise,
       ef.periode                                                                                          AS date_effectif,
       n.libelle_n5,
       n.libelle_n1,
       et.code_activite,
       COALESCE(ep.last_procol, 'in_bonis'::text)                                                          AS last_procol,
       ep.date_effet                                                                                       AS date_last_procol,
       COALESCE(ap.ap, false)                                                                              AS activite_partielle,
       ap.heure_consomme                                                                                   AS apconso_heure_consomme,
       ap.montant                                                                                          AS apconso_montant,
       u.dette[1] > u.dette[2] OR u.dette[2] > u.dette[3]                                                  AS hausse_urssaf,
       u.dette[1]                                                                                          AS dette_urssaf,
       u.periode_urssaf,
       u.presence_part_salariale,
       s.alert,
       et.siege,
       g.raison_sociale                                                                                    AS raison_sociale_groupe,
       ti.code_commune IS NOT NULL                                                                         AS territoire_industrie,
       ti.code_terrind                                                                                     AS code_territoire_industrie,
       ti.libelle_terrind                                                                                  AS libelle_territoire_industrie,
       cj.libelle                                                                                          AS statut_juridique_n3,
       cj2.libelle                                                                                         AS statut_juridique_n2,
       cj1.libelle                                                                                         AS statut_juridique_n1,
       et.creation                                                                                         AS date_ouverture_etablissement,
       en.creation                                                                                         AS date_creation_entreprise,
       COALESCE(sco.secteur_covid, 'nonSecteurCovid'::text)                                                AS secteur_covid,
       CASE
         WHEN COALESCE(en.etat_administratif, 'A'::text) = 'C'::text THEN 'F'::text
         ELSE COALESCE(et.etat_administratif, 'A'::text)
         END                                                                                             AS etat_administratif,
       en.etat_administratif                                                                               AS etat_administratif_entreprise
FROM last_liste l
       JOIN etablissement0 et ON true
       JOIN entreprise0 en ON en.siren::text = et.siren::text
       JOIN v_etablissement_raison_sociale etrs ON etrs.id_etablissement = et.id
       JOIN v_roles r ON et.siren::text = r.siren::text
       JOIN departements d ON d.code = et.departement
       LEFT JOIN v_naf n ON n.code_n5 = et.code_activite
       LEFT JOIN secteurs_covid sco ON sco.code_activite = et.code_activite
       LEFT JOIN score0 s ON et.siret::text = s.siret AND s.libelle_liste = l.last_liste
       LEFT JOIN v_alert_etablissement aet ON aet.siret = et.siret::text
       LEFT JOIN v_alert_entreprise aen ON aen.siren = et.siren::text
       LEFT JOIN v_last_effectif ef ON ef.siret::text = et.siret::text
       LEFT JOIN v_hausse_urssaf u ON u.siret::text = et.siret::text
       LEFT JOIN v_apdemande ap ON ap.siret::text = et.siret::text
       LEFT JOIN v_last_procol ep ON ep.siren::text = et.siren::text
       LEFT JOIN v_diane_variation_ca di ON di.siren::text = et.siren::text
       LEFT JOIN entreprise_ellisphere0 g ON g.siren::text = et.siren::text
       LEFT JOIN terrind ti ON ti.code_commune = et.code_commune
       LEFT JOIN categorie_juridique cj ON cj.code = en.statut_juridique
       LEFT JOIN categorie_juridique cj2 ON "substring"(cj.code, 0, 3) = cj2.code
       LEFT JOIN categorie_juridique cj1 ON "substring"(cj.code, 0, 2) = cj1.code;

alter materialized view v_summaries owner to postgres;

create index idx_v_summaries_order_raison_sociale
  on v_summaries (raison_sociale, siret);

create index idx_v_summaries_raison_sociale
  on v_summaries using gin (siret gin_trgm_ops, raison_sociale gin_trgm_ops);

create index idx_v_summaries_roles
  on v_summaries using gin (roles);

create index idx_v_summaries_score
  on v_summaries (valeur_score desc, siret asc, siege asc, effectif asc, effectif_entreprise asc, code_departement asc,
                  last_procol asc, chiffre_affaire asc, etat_administratif asc, first_alert asc, first_list_entreprise
                  asc, first_red_list_entreprise asc)
  where (alert <> 'Pas d''alerte'::text);

create index idx_v_summaries_siren
  on v_summaries (siren);

create index idx_v_summaries_siret
  on v_summaries (siret);

