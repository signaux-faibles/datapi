truncate table etablissement_apdemande;
insert into etablissement_apdemande (siret, siren, id_demande, effectif_entreprise, effectif, date_statut,
                                     periode_start, periode_end, hta, mta, effectif_autorise, motif_recours_se,
                                     heure_consomme, montant_consomme, effectif_consomme)
select etab_siret,
       substring(etab_siret from 0 for 9),
       id_da,
       eff_ent,
       eff_etab,
       date_statut::date,
       date_deb::date,
       date_fin::date,
       hta,
       mta,
       eff_auto,
       motif_recours_se,
       s_heure_consom_tot,
       s_montant_consom_tot,
       s_eff_consom_tot
from demande_ap;

truncate table etablissement_apconso;
insert into etablissement_apconso (siret, siren, id_conso, heure_consomme, montant, effectif, periode)
select etab_siret, substring(etab_siret from 1 for 9), id_da, heures, montants, effectifs, mois::date
from consommation_ap;
