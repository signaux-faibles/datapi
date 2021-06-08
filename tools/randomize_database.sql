-- Nettoyage des etablissements et/ou entreprises orphelines
drop table if exists todelete;
create temporary table todelete as
select en.siren as siren from entreprise0 en
left join etablissement0 et on et.siren = en.siren
where et.siren is null;

insert into todelete (siren)
select distinct et.siren from entreprise0 en
right join etablissement0 et on et.siren = en.siren
where en.siren is null;

delete from etablissement where siren in (select siren from todelete);
delete from etablissement_apconso where siren in (select siren from todelete);
delete from etablissement_apdemande where siren in (select siren from todelete);
delete from etablissement_delai where siren in (select siren from todelete);
delete from etablissement_follow where siren in (select siren from todelete);
delete from etablissement_periode_urssaf where siren in (select siren from todelete);
delete from etablissement_procol where siren in (select siren from todelete);
delete from entreprise where siren in (select siren from todelete);
delete from entreprise_bdf where siren in (select siren from todelete);
delete from entreprise_diane where siren in (select siren from todelete);
delete from entreprise_ellisphere where siren in (select siren from todelete);
delete from entreprise_paydex where siren in (select siren from todelete);

-- generation de nouveaux sirets aléatoires
create table translate_siret as 
select
  e.siret, string_agg(substr(characters, (random() * (length(characters) - 1)+1)::integer, 1), '') as new_siret
from (values('0123456789')) as symbols(characters)
  join generate_series(1, 14) on  true
  join etablissement e on true
group by siret;

-- conservation du groupement des établissements par entreprise
with t as (select substring(siret from 1 for 9) siren, 
		   array_agg(new_siret) as old_sirets, 
		   first(substring(new_siret from 1 for 9)) new_siren
 from translate_siret
 group by substring(siret from 1 for 9)
 ),
u as (select siren, unnest(old_sirets) as old_siret, new_siren || substring(unnest(old_sirets) from 10 for 5) new_siret 
	  from t)
update translate_siret t set new_siret = u.new_siret
from u where u.old_siret = t.new_siret;

-- updates aléatoires
with seeda as (
select
  e.id, 
  string_agg(substr(characters, (random() * (length(characters) - 1)+1)::integer, 1), '') as seed1,
  string_agg(substr(characters, (random() * (length(characters) - 1)+1)::integer, 1), '') as seed2,
  string_agg(substr(characters, (random() * (length(characters) - 1)+1)::integer, 1), '') as seed3,
  string_agg(substr(characters, (random() * (length(characters) - 1)+1)::integer, 1), '') as seed4
from (values('ABCDEFGHIJKLMNOPQRSTUVWXYZ')) as symbols(characters)
  join generate_series(1, 20) on  true
  inner join etablissement e on true
group by e.id),
seedi as (
select
  e.id, 
  lpad((random() * 94000 + 1000)::integer::text, 5, '0') as code_postal
from (values('0123456789')) as symbols(characters)
  join generate_series(1, 5) on  true
  inner join etablissement e on true
group by e.id)
update etablissement e set siret = t.new_siret, siren = substring(t.new_siret from 1 for 9),
complement_adresse = (select complement_adresse from etablissement order by random() limit 1),
numero_voie = (random()*98+1)::integer,
indice_repetition = (select indice_repetition from etablissement order by random() limit 1),
type_voie = 'RUE',
voie = substring(a.seed1 from 1 for 6),
commune_etranger = null,
code_cedex = null,
code_pays_etranger = null,
pays_etranger = null, 
code_postal = i.code_postal,
departement = substring(i.code_postal from 1 for 2),
latitude = latitude * random()*0.02+0.99,
longitude = longitude * random()*0.02+0.99 
from translate_siret t 
inner join seeda a on true
inner join seedi i on true
where t.siret = e.siret and a.id = e.id and i.id = e.id;

update etablissement_apconso e set siret = t.new_siret, siren = substring(t.new_siret from 1 for 9)
from translate_siret t where t.siret = e.siret;

update etablissement_apdemande e set siret = t.new_siret, siren = substring(t.new_siret from 1 for 9)
from translate_siret t where t.siret = e.siret;

update etablissement_procol e set siret = t.new_siret, siren = substring(t.new_siret from 1 for 9)
from translate_siret t where t.siret = e.siret;

with t as (select distinct substring(siret from 1 for 9) as siren, 
		  substring(new_siret from 1 for 9) as new_siren 
		  from translate_siret)
update entreprise e set siren = t.new_siren
from t where t.siren = e.siren;

with raisoc as (
select
  e.id, 
  string_agg(substr(characters, (random() * (length(characters) - 1)+1)::integer, 1), '') as raison_sociale
from (values('ABCDEFGHIJKLMNOPQRSTUVWXYZ')) as symbols(characters)
  join generate_series(1, 20) on  true
  inner join entreprise e on true
group by e.id)
update entreprise e set 
  raison_sociale = r.raison_sociale, 
  prenom1 = '',
  prenom2 = '',
  prenom3 = '',
  prenom4 = '',
  nom = '',
  nom_usage = ''
from raisoc r where r.id = e.id;

with t as (select distinct substring(siret from 1 for 9) as siren, 
		  substring(new_siret from 1 for 9) as new_siren
		  from translate_siret),
seed as (
select
  e.id, 
  string_agg(substr(characters, (random() * (length(characters) - 1)+1)::integer, 1), '') as seed1,
  string_agg(substr(characters, (random() * (length(characters) - 1)+1)::integer, 1), '') as seed2,
  string_agg(substr(characters, (random() * (length(characters) - 1)+1)::integer, 1), '') as seed3,
  string_agg(substr(characters, (random() * (length(characters) - 1)+1)::integer, 1), '') as seed4,
  string_agg(substr(characters, (random() * (length(characters) - 1)+1)::integer, 1), '') as seed5,
  string_agg(substr(characters, (random() * (length(characters) - 1)+1)::integer, 1), '') as seed6
from (values('ABCDEFGHIJKLMNOPQRSTUVWXYZ')) as symbols(characters)
  join generate_series(1, 10) on  true
  inner join entreprise_ellisphere e on true
group by e.id)
update entreprise_ellisphere e set siren = t.new_siren,
code = s.seed1,
refid = s.seed2,
raison_sociale = s.seed3,
adresse = s.seed4,
part_financiere = round((random() * 100)::numeric, 2),
code_filiere = s.seed5,
refid_filiere = s.seed6
from t 
inner join seed s on true
where t.siren = e.siren and s.id = e.id;

with t as (select distinct substring(siret from 1 for 9) as siren, 
		  substring(new_siret from 1 for 9) as new_siren 
		  from translate_siret)
update entreprise_bdf e set siren = t.new_siren,
  arrete_bilan_bdf = arrete_bilan_bdf + (random()*(6)-3)::integer * '1 month'::interval,
  delai_fournisseur = delai_fournisseur * random() * 2,
  financier_court_terme = financier_court_terme * random() * 2,
  poids_frng = poids_frng * random()*2,
  dette_fiscale = poids_frng * random()*2,
  frais_financier = frais_financier * random()*2 + 10*random(),
  taux_marge = taux_marge * random()*4-2
from t where t.siren = e.siren;

with t as (select distinct substring(siret from 1 for 9) as siren, 
		  substring(new_siret from 1 for 9) as new_siren 
		  from translate_siret)
update entreprise_paydex e set siren = t.new_siren,
  nb_jours = nb_jours * random()*2
from t where t.siren = e.siren;

with t as (select distinct substring(siret from 1 for 9) as siren, 
		  substring(new_siret from 1 for 9) as new_siren
		  from translate_siret)
update entreprise_diane e set siren = t.new_siren,
  arrete_bilan_diane = arrete_bilan_diane + (random()*(6)-3)::integer * '1 month'::interval,
  achat_marchandises = round((achat_marchandises * (random()*(10-0.1)+0.1))::numeric),
  achat_matieres_premieres = round((achat_matieres_premieres * (random()*(10-0.1)+0.1))::numeric),
  autonomie_financiere = round((autonomie_financiere * (random()*(4)-2))::numeric, 2),
  autres_achats_charges_externes = round((autres_achats_charges_externes * (random()*(10-0.1)+0.1))::numeric),
  autres_produits_charges_reprises = round((autres_produits_charges_reprises * (random()*(10-0.1)+0.1))::numeric),
  benefice_ou_perte = round((benefice_ou_perte * (random()*(4)-2))::numeric, 2),
  ca_exportation = ca_exportation * round((random()*(10-0.1)+0.1)::numeric),
  capacite_autofinancement = round((capacite_autofinancement * (random()*(4)-2))::numeric, 2),
  capacite_remboursement = round((capacite_remboursement * (random()*(4)-2))::numeric, 2),
  ca_par_effectif = round((ca_par_effectif * (random()*(4)-2))::numeric, 2),
  charge_exceptionnelle = round((charge_exceptionnelle * (random()*(10-0.1)+0.1))::numeric),
  charge_personnel = round((charge_personnel * (random()*(10-0.1)+0.1))::numeric),
  charges_financieres = round((charges_financieres * (random()*(10-0.1)+0.1))::numeric),
  chiffre_affaire = round((chiffre_affaire * (random()*(10-0.1)+0.1))::numeric) ,
  conces_brev_et_droits_sim = round((conces_brev_et_droits_sim * (random()*(10-0.1)+0.1))::numeric),
  consommation = round((consommation * (random()*(10-0.1)+0.1))::numeric),
  couverture_ca_besoin_fdr = round((couverture_ca_besoin_fdr * (random()*(4)-2))::numeric, 2),
  couverture_ca_fdr = round((couverture_ca_fdr * (random()*(4)-2))::numeric, 2),
  credit_client = round((credit_client * (random()*(10-0.1)+0.1))::numeric),
  credit_fournisseur = round((credit_fournisseur * (random()*(4)-2))::numeric),
  degre_immo_corporelle = round((degre_immo_corporelle * (random()*(4)-2))::numeric, 2),
  dette_fiscale_et_sociale = round((dette_fiscale_et_sociale * (random()*(4)-2))::numeric, 2),
  dotation_amortissement = round((dotation_amortissement * (random()*(10-0.1)+0.1))::numeric, 2),
  effectif_consolide = round((effectif_consolide * (random()*(10-0.1)+0.1))::integer, 2),
  endettement = round((endettement * (random()*(4)-2))::numeric, 2),
  endettement_global = round((endettement_global * (random()*(4)-2))::numeric, 2),
  equilibre_financier = round((equilibre_financier * (random()*(4)-2))::numeric, 2),
  excedent_brut_d_exploitation = round((excedent_brut_d_exploitation * (random()*(10-0.1)+0.1))::numeric),
  exportation = round((exportation * (random()*(10-0.1)+0.1))::numeric),
  financement_actif_circulant = round((financement_actif_circulant * (random()*(4)-2))::numeric, 2),
  frais_de_RetD = round((frais_de_RetD * (random()*(10-0.1)+0.1))::numeric, 2),
  impot_benefice = round((impot_benefice * (random()*(10-0.1)+0.1))::numeric, 2),
  impots_taxes = round((impots_taxes * (random()*(10-0.1)+0.1))::numeric),
  independance_financiere = round((independance_financiere * (random()*(4)-2))::numeric,2),
  interets = round((interets * (random()*(10-0.1)+0.1))::numeric),
  liquidite_generale = round((liquidite_generale * (random()*(4)-2))::numeric,2),
  liquidite_reduite = round((liquidite_reduite * (random()*(4)-2))::numeric, 2),
  marge_commerciale = round((marge_commerciale * (random()*(4)-2))::numeric, 2),
  nombre_etab_secondaire = (nombre_etab_secondaire * random() * 2)::integer,
  nombre_filiale = nombre_filiale + random()*(3)::integer,
  nombre_mois = nombre_mois + random()*(3)::integer,
  operations_commun = round((operations_commun * (random()*(10-0.1)+0.1))::numeric),
  part_autofinancement = round((part_autofinancement * (random()*(4)-2))::numeric, 2),
  part_etat = round((part_etat * (random()*(4)-2))::numeric, 2),
  part_preteur = round((part_preteur * (random()*(4)-2))::numeric, 2),
  part_salaries = round((part_salaries * (random()*(4)-2))::numeric, 2),
  participation_salaries = round((participation_salaries * (random()*(10-0.1)+0.1))::numeric),
  performance = round((performance * (random()*(4)-2))::numeric, 2),
  poids_bfr_exploitation = round((poids_bfr_exploitation * (random()*(10-0.1)+0.1))::numeric),
  production = round((production * (random()*(10-0.1)+0.1))::numeric),
  productivite_capital_financier = round((productivite_capital_financier * (random()*(4)-2))::numeric, 2),
  productivite_capital_investi = round((productivite_capital_investi * (random()*(4)-2))::numeric, 2),
  productivite_potentiel_production = round((productivite_potentiel_production * (random()*(4)-2))::numeric, 2),
  produit_exceptionnel = round((produit_exceptionnel * (random()*(10-0.1)+0.1))::numeric),
  produits_financiers = round((produits_financiers * (random()*(10-0.1)+0.1))::numeric),
  rendement_brut_fonds_propres = round((rendement_brut_fonds_propres * (random()*(4)-2))::numeric,2),
  rendement_capitaux_propres = round((rendement_capitaux_propres * (random()*(4)-2))::numeric, 2),
  rendement_ressources_durables = round((rendement_ressources_durables * (random()*(4)-2))::numeric, 2),
  rentabilite_economique = round((rentabilite_economique * (random()*(4)-2))::numeric, 2),
  rentabilite_nette = round((rentabilite_nette * (random()*(4)-2))::numeric, 2),
  resultat_avant_impot = round((resultat_avant_impot * (random()*(10-0.1)+0.1))::numeric),
  resultat_expl = round((resultat_expl * (random()*(10-0.1)+0.1))::numeric),
  rotation_stocks = round((rotation_stocks * (random()*(4)-2))::numeric, 2),
  subventions_d_exploitation = round((subventions_d_exploitation * (random()*(10-0.1)+0.1))::numeric),
  taille_compo_groupe = taille_compo_groupe + (random()*(4))::integer,
  taux_d_investissement_productif = round((taux_d_investissement_productif * (random()*(4)-2))::numeric, 2),
  taux_endettement = round((taux_endettement * (random()*(4)-2))::numeric, 2),
  taux_interet_financier = round((taux_interet_financier * (random()*(4)-2))::numeric, 2),
  taux_interet_sur_ca = round((taux_interet_sur_ca * (random()*(4)-2))::numeric, 2),
  taux_valeur_ajoutee = round((taux_valeur_ajoutee * (random()*(4)-2))::numeric, 2),
  valeur_ajoutee = round((valeur_ajoutee * (random()*(10-0.1)+0.1))::numeric)
from t where t.siren = e.siren;

-- activite partielle
create table translate_ap as 
select
  substring(e.id_demande from 1 for 9) as id_demande, 
  string_agg(substr(characters, (random() * (length(characters) - 1)+1)::integer, 1), '') as new_demande
from (values('0123456789')) as symbols(characters)
  join generate_series(1, 9) on  true
  join (select distinct substring(id_demande from 1 for 9) id_demande from etablissement_apdemande) e on true
group by substring(e.id_demande from 1 for 9);

update etablissement_apdemande e set id_demande = t.new_demande || substring(e.id_demande from 10 for 2),
effectif_entreprise=round((effectif_entreprise*(random()*(10-0.1)+0.1))::numeric),
effectif=round((effectif * (random()*(10-0.1)+0.1))::numeric),
hta=hta * (random()*(10-0.1)+0.1),
mta=mta * (random()*(10-0.1)+0.1),
effectif_autorise=round((effectif_autorise * (random()*(10-0.1)+0.1))::numeric),
heure_consomme=round((heure_consomme * (random()*(10-0.1)+0.1))::numeric, 2),
montant_consomme=round((montant_consomme * (random()*(10-0.1)+0.1))::numeric, 2),
effectif_consomme=round((effectif_consomme * (random()*(10-0.1)+0.1))::numeric)
from translate_ap t where substring(e.id_demande from 1 for 9) = t.id_demande;

update etablissement_apconso e set id_conso = a.new_demande || substring(e.id_conso from 10 for 2), 
heure_consomme = heure_consomme * (random()*(10-0.1)+0.1),
montant = round((montant * (random()*(10-0.1)+0.1))::numeric,2),
effectif = round((effectif * (random()*(10-0.1)+0.1))::numeric)
from translate_ap a where a.id_demande = substring(id_conso from 1 for 9);

-- urssaf
create table translate_compte as 
select
  numero_compte, 
  string_agg(substr(characters, (random() * (length(characters) - 1)+1)::integer, 1), '') as new_compte
from (values('0123456789')) as symbols(characters)
  join generate_series(1, 18) on  true
  join (select distinct numero_compte from etablissement_delai) e on true
group by numero_compte;

update etablissement_delai e set numero_compte = t.new_compte
from translate_compte t where t.numero_compte = e.numero_compte;

update etablissement_delai e set siret = t.new_siret,
siren = substring(t.new_siret from 1 for 9),
montant_echeancier = montant_echeancier * (random()*(10-0.1)+0.1),
numero_contentieux = 'absent',
denomination = 'absent'
from translate_siret t
where t.siret = e.siret;

update etablissement_periode_urssaf e set siret = t.new_siret,
siren = substring(t.new_siret from 1 for 9),
cotisation = round((cotisation * (random()*(10-0.1)+0.1))::numeric,2),
part_patronale = round((part_patronale * (random()*(10-0.1)+0.1))::numeric,2),
part_salariale = round((part_salariale * (random()*(10-0.1)+0.1))::numeric,2),
effectif = round((effectif * (random()*(10-0.1)+0.1))::numeric)
from translate_siret t
where t.siret = e.siret;

update score e set siret = t.new_siret,
siren = substring(t.new_siret from 1 for 9),
score = greatest(score * random()*2, 0.99)
from translate_siret t
where t.siret = e.siret;

refresh materialized view v_alert_entreprise;
refresh materialized view v_alert_etablissement;
refresh materialized view v_apdemande;
refresh materialized view v_diane_variation_ca;
refresh materialized view v_etablissement_raison_sociale;
refresh materialized view v_hausse_urssaf;
refresh materialized view v_last_effectif;
refresh materialized view v_last_procol;
refresh materialized view v_naf;
refresh materialized view v_roles;
refresh materialized view v_summaries;

drop table todelete;
drop table translate_siret;
drop table translate_ap;
drop table translate_compte;