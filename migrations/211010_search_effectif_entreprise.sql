create or replace view etablissement0 as
select * from etablissement where version = 0;

create or replace view entreprise0 as
select * from entreprise where version = 0;

create table secteurs_covid as
with js as (select '{
  "s1": [
    {
      "codeActivite": "4939C",
      "libelleActivite": "Téléphériques et remontées mécaniques"
    },
    {
      "codeActivite": "5510Z",
      "libelleActivite": "Hôtels et hébergement similaire"
    },
    {
      "codeActivite": "5520Z",
      "libelleActivite": "Hébergement touristique et autre hébergement de courte durée"
    },
    {
      "codeActivite": "5530Z",
      "libelleActivite": "Terrains de camping et parcs pour caravanes ou véhicules de loisirs"
    },
    {
      "codeActivite": "5610A",
      "libelleActivite": "Restauration traditionnelle"
    },
    {
      "codeActivite": "5610B",
      "libelleActivite": "Cafétérias et autres libres-services"
    },
    {
      "codeActivite": "5610C",
      "libelleActivite": "Restauration de type rapide"
    },
    {
      "codeActivite": "5629A",
      "libelleActivite": "Services de restauration collective sous contrat, de cantines et restaurants d''entreprise"
    },
    {
      "codeActivite": "5629B",
      "libelleActivite": "Services de restauration collective sous contrat, de cantines et restaurants d''entreprise"
    },
    {
      "codeActivite": "5621Z",
      "libelleActivite": "Services des traiteurs"
    },
    {
      "codeActivite": "5630Z",
      "libelleActivite": "Débits de boissons"
    },
    {
      "codeActivite": "5914Z",
      "libelleActivite": "Projection de films cinématographiques et autres industries techniques du cinéma et de l''image animée"
    },
    {
      "codeActivite": "5912Z",
      "libelleActivite": "Post-production de films cinématographiques, de vidéo et de programmes de télévision"
    },
    {
      "codeActivite": "5913A",
      "libelleActivite": "Distribution de films cinématographiques"
    },
    {
      "codeActivite": "7721Z",
      "libelleActivite": "Location et location-bail d''articles de loisirs et de sport"
    },
    {
      "codeActivite": "7911Z",
      "libelleActivite": "Activités des agences de voyage"
    },
    {
      "codeActivite": "7912Z",
      "libelleActivite": "Activités des voyagistes"
    },
    {
      "codeActivite": "7990Z",
      "libelleActivite": "Autres services de réservation et activités connexes"
    },
    {
      "codeActivite": "8230Z",
      "libelleActivite": "Organisation de foires, évènements publics ou privés, salons ou séminaires professionnels, congrès"
    },
    {
      "codeActivite": "8551Z",
      "libelleActivite": "Enseignement de disciplines sportives et d''activités de loisirs"
    },
    {
      "codeActivite": "9001Z",
      "libelleActivite": "Arts du spectacle vivant, cirques"
    },
    {
      "codeActivite": "9002Z",
      "libelleActivite": "Activités de soutien au spectacle vivant"
    },
    {
      "codeActivite": "9003A",
      "libelleActivite": "Création artistique relevant des arts plastiques"
    },
    {
      "codeActivite": "9003B",
      "libelleActivite": "Artistes auteurs"
    },
    {
      "codeActivite": "9004Z",
      "libelleActivite": "Gestion de salles de spectacles et production de spectacles"
    },
    {
      "codeActivite": "9102Z",
      "libelleActivite": "Gestion des musées"
    },
    {
      "codeActivite": "9103Z",
      "libelleActivite": "Gestion des sites et monuments historiques et des attractions touristiques similaires"
    },
    {
      "codeActivite": "9104Z",
      "libelleActivite": "Gestion des jardins botaniques et zoologiques et des réserves naturelles"
    },
    {
      "codeActivite": "9311Z",
      "libelleActivite": "Gestion d''installations sportives"
    },
    {
      "codeActivite": "9312Z",
      "libelleActivite": "Activités de clubs de sports"
    },
    {
      "codeActivite": "9313Z",
      "libelleActivite": "Activité des centres de culture physique"
    },
    {
      "codeActivite": "9319Z",
      "libelleActivite": "Autres activités liées au sport"
    },
    {
      "codeActivite": "9321Z",
      "libelleActivite": "Activités des parcs d''attractions et parcs à thèmes, fêtes foraines"
    },
    {
      "codeActivite": "9329Z",
      "libelleActivite": "Autres activités récréatives et de loisirs"
    },
    {
      "codeActivite": "9604Z",
      "libelleActivite": "Entretien corporel"
    },
    {
      "codeActivite": "5110Z",
      "libelleActivite": "Transport aérien de passagers"
    },
    {
      "codeActivite": "4939A",
      "libelleActivite": "Transports routiers réguliers de voyageurs"
    },
    {
      "codeActivite": "4939B",
      "libelleActivite": "Autres transports routiers de voyageurs"
    },
    {
      "codeActivite": "5010Z",
      "libelleActivite": "Transport maritime et côtier de passagers"
    },
    {
      "codeActivite": "5911A",
      "libelleActivite": "Production de films et de programmes pour la télévision"
    },
    {
      "codeActivite": "5911B",
      "libelleActivite": "Production de films institutionnels et publicitaires"
    },
    {
      "codeActivite": "5911C",
      "libelleActivite": "Production de films pour le cinéma"
    },
    {
      "codeActivite": "7420Z",
      "libelleActivite": "Activités photographiques"
    },
    {
      "codeActivite": "8552Z",
      "libelleActivite": "Enseignement culturel"
    },
    {
      "codeActivite": "7430Z",
      "libelleActivite": "Traducteurs-interprètes"
    },
    {
      "codeActivite": "4932Z",
      "libelleActivite": "Transports de voyageurs par taxis et véhicules de tourisme avec chauffeur"
    },
    {
      "codeActivite": "7711A",
      "libelleActivite": "Location de courte durée de voitures et de véhicules automobiles légers"
    },
    {
      "codeActivite": "2511Z",
      "libelleActivite": "Fabrication de structures métalliques et de parties de structures"
    },
    {
      "codeActivite": "7312Z",
      "libelleActivite": "Régie publicitaire de médias"
    },
    {
      "codeActivite": "0127Z",
      "libelleActivite": "Culture de plantes à boissons"
    },
    {
      "codeActivite": "0121Z",
      "libelleActivite": "Culture de la vigne"
    },
    {
      "codeActivite": "1101Z",
      "libelleActivite": "Production de boissons alcooliques distillées"
    },
    {
      "codeActivite": "1102A",
      "libelleActivite": "Fabrication de vins effervescents"
    },
    {
      "codeActivite": "1102B",
      "libelleActivite": "Vinification"
    },
    {
      "codeActivite": "1103Z",
      "libelleActivite": "Fabrication de cidre et de vins de fruits"
    },
    {
      "codeActivite": "1104Z",
      "libelleActivite": "Production d''autres boissons fermentées non distillées"
    }
  ],
  "s1Possible": [
    {
      "codeActivite": "6612Z",
      "libelleActivite": "Entreprises de détaxe et bureaux de change (changeurs manuels)"
    },
    {
      "codeActivite": "9200Z",
      "libelleActivite": "Exploitations de casinos"
    },
    {
      "codeActivite": "4910Z",
      "libelleActivite": "Trains et chemins de fer touristiques"
    },
    {
      "codeActivite": "5030Z",
      "libelleActivite": "Transport de passagers sur les fleuves, les canaux, les lacs, location de bateaux de plaisance"
    }
  ],
  "s1bis": [
    {
      "codeActivite": "0311Z",
      "libelleActivite": "Pêche en mer"
    },
    {
      "codeActivite": "0312Z",
      "libelleActivite": "Pêche en eau douce"
    },
    {
      "codeActivite": "0321Z",
      "libelleActivite": "Aquaculture en mer"
    },
    {
      "codeActivite": "0322Z",
      "libelleActivite": "Aquaculture en eau douce"
    },
    {
      "codeActivite": "1105Z",
      "libelleActivite": "Fabrication de bière"
    },
    {
      "codeActivite": "1106Z",
      "libelleActivite": "Fabrication de malt"
    },
    {
      "codeActivite": "4617A",
      "libelleActivite": "Centrales d''achat alimentaires"
    },
    {
      "codeActivite": "4631Z",
      "libelleActivite": "Commerce de gros de fruits et légumes"
    },
    {
      "codeActivite": "0130Z",
      "libelleActivite": "Herboristerie/ horticulture/ commerce de gros de fleurs et plans"
    },
    {
      "codeActivite": "4622Z",
      "libelleActivite": "Herboristerie/ horticulture/ commerce de gros de fleurs et plans"
    },
    {
      "codeActivite": "4633Z",
      "libelleActivite": "Commerce de gros de produits laitiers, œufs, huiles et matières grasses comestibles"
    },
    {
      "codeActivite": "4634Z",
      "libelleActivite": "Commerce de gros de boissons"
    },
    {
      "codeActivite": "4638A",
      "libelleActivite": "Mareyage et commerce de gros de poissons, coquillages, crustacés"
    },
    {
      "codeActivite": "4638B",
      "libelleActivite": "Commerce de gros alimentaire spécialisé divers"
    },
    {
      "codeActivite": "4639A",
      "libelleActivite": "Commerce de gros de produits surgelés"
    },
    {
      "codeActivite": "4639B",
      "libelleActivite": "Commerce de gros alimentaire"
    },
    {
      "codeActivite": "4690Z",
      "libelleActivite": "Commerce de gros non spécialisé"
    },
    {
      "codeActivite": "4641Z",
      "libelleActivite": "Commerce de gros de textiles"
    },
    {
      "codeActivite": "4618Z",
      "libelleActivite": "Intermédiaires spécialisés dans le commerce d''autres produits spécifiques"
    },
    {
      "codeActivite": "4642Z",
      "libelleActivite": "Commerce de gros d''habillement et de chaussures"
    },
    {
      "codeActivite": "4649Z",
      "libelleActivite": "Commerce de gros d''autres biens domestiques"
    },
    {
      "codeActivite": "4644Z",
      "libelleActivite": "Commerce de gros de vaisselle, verrerie et produits d''entretien"
    },
    {
      "codeActivite": "4669C",
      "libelleActivite": "Commerce de gros de fournitures et équipements divers pour le commerce et les services"
    },
    {
      "codeActivite": "9601A",
      "libelleActivite": "Blanchisserie-teinturerie de gros"
    },
    {
      "codeActivite": "5920Z",
      "libelleActivite": "Enregistrement sonore et édition musicale"
    },
    {
      "codeActivite": "5811Z",
      "libelleActivite": "Editeurs de livres (Édition de livres)"
    },
    {
      "codeActivite": "5223Z",
      "libelleActivite": "Services auxiliaires des transports aériens"
    },
    {
      "codeActivite": "5222Z",
      "libelleActivite": "Services auxiliaires de transport par eau"
    },
    {
      "codeActivite": "8010Z",
      "libelleActivite": "Activités de sécurité privée"
    },
    {
      "codeActivite": "8121Z",
      "libelleActivite": "Nettoyage courant des bâtiments"
    },
    {
      "codeActivite": "8122Z",
      "libelleActivite": "Autres activités de nettoyage des bâtiments et nettoyage industriel"
    },
    {
      "codeActivite": "1013B",
      "libelleActivite": "Préparation à caractère artisanal de produits de charcuterie"
    },
    {
      "codeActivite": "1071D",
      "libelleActivite": "Pâtisserie"
    },
    {
      "codeActivite": "4722Z",
      "libelleActivite": "Commerce de détail de viandes et de produits à base de viande en magasin spécialisé"
    },
    {
      "codeActivite": "1412Z",
      "libelleActivite": "Fabrication de vêtements de travail"
    },
    {
      "codeActivite": "1820Z",
      "libelleActivite": "Reproduction d''enregistrements"
    },
    {
      "codeActivite": "2313Z",
      "libelleActivite": "Fabrication de verre creux"
    },
    {
      "codeActivite": "2341Z",
      "libelleActivite": "Fabrication d''articles céramiques à usage domestique ou ornemental"
    },
    {
      "codeActivite": "2571Z",
      "libelleActivite": "Fabrication de coutellerie"
    },
    {
      "codeActivite": "2599A",
      "libelleActivite": "Fabrication d''articles métalliques ménagers"
    },
    {
      "codeActivite": "2752Z",
      "libelleActivite": "Fabrication d''appareils ménagers non électriques"
    },
    {
      "codeActivite": "2740Z",
      "libelleActivite": "Fabrication d''appareils d''éclairage électrique"
    },
    {
      "codeActivite": "4321A",
      "libelleActivite": "Travaux d''installation électrique dans tous locaux"
    },
    {
      "codeActivite": "4332C",
      "libelleActivite": "Aménagement (Agencement) de lieux de vente"
    },
    {
      "codeActivite": "7021Z",
      "libelleActivite": "Conseil en relations publiques et communication"
    },
    {
      "codeActivite": "7311Z",
      "libelleActivite": "Activités des agences de publicité"
    },
    {
      "codeActivite": "7410Z",
      "libelleActivite": "Activités spécialisées de design"
    },
    {
      "codeActivite": "7490B",
      "libelleActivite": "Activités spécialisées, scientifiques et techniques diverses"
    },
    {
      "codeActivite": "9601B",
      "libelleActivite": "Blanchisserie-teinturerie de détail"
    },
    {
      "codeActivite": "4799B",
      "libelleActivite": "Vente par automate"
    },
    {
      "codeActivite": "4632B",
      "libelleActivite": "Commerce de gros de viandes et de produits à base de viande"
    },
    {
      "codeActivite": "7810Z",
      "libelleActivite": "Activités des agences de placement de main-d''œuvre"
    },
    {
      "codeActivite": "1413Z",
      "libelleActivite": "Couturiers"
    },
    {
      "codeActivite": "9523Z",
      "libelleActivite": "Réparation de chaussures et d''articles en cuir"
    }
  ],
  "s1bisPossible": [
    {
      "codeActivite": "1051C",
      "libelleActivite": "Production de fromages sous appellation d''origine protégée ou indication géographique protégée"
    },
    {
      "codeActivite": "4617B",
      "libelleActivite": "Autres intermédiaires du commerce en denrées et boissons"
    },
    {
      "codeActivite": "4730Z",
      "libelleActivite": "Stations-service"
    },
    {
      "codeActivite": "1013A",
      "libelleActivite": "Fabrication de foie gras"
    },
    {
      "codeActivite": "4781Z",
      "libelleActivite": "Commerce de détail de viande, produits à base de viandes sur éventaires et marchés"
    },
    {
      "codeActivite": "4789Z",
      "libelleActivite": "Commerce de détail de livres sur éventaires et marchés"
    },
    {
      "codeActivite": "6622Z",
      "libelleActivite": "Courtier en assurance voyage"
    },
    {
      "codeActivite": "6820B",
      "libelleActivite": "Location et exploitation d''immeubles non résidentiels de réception"
    },
    {
      "codeActivite": "8211Z",
      "libelleActivite": "Services administratifs d''assistance à la demande de visas"
    },
    {
      "codeActivite": "4120A",
      "libelleActivite": "Construction de maisons mobiles pour les terrains de camping"
    },
    {
      "codeActivite": "9609Z",
      "libelleActivite": "Garde d''animaux de compagnie avec ou sans hébergement"
    },
    {
      "codeActivite": "1399Z",
      "libelleActivite": "Fabrication de dentelle et broderie"
    },
    {
      "codeActivite": "6010Z",
      "libelleActivite": "Edition et diffusion de programmes radios à audience locale, éditions de chaînes de télévision à audience locale"
    },
    {
      "codeActivite": "6020B",
      "libelleActivite": "Edition et diffusion de programmes radios à audience locale, éditions de chaînes de télévision à audience locale"
    },
    {
      "codeActivite": "3230Z",
      "libelleActivite": "Fabrication de skis, fixations et bâtons pour skis, chaussures de ski"
    },
    {
      "codeActivite": "2591Z",
      "libelleActivite": "Fabrication de bidons de bière métalliques, tonnelets de bière métalliques, fûts de bière métalliques"
    }
  ],
  "s2": [
    {
      "codeActivite": "4511Z",
      "libelleActivite": "Commerce de voitures et de véhicules automobiles légers"
    },
    {
      "codeActivite": "4719A",
      "libelleActivite": "Grands magasins"
    },
    {
      "codeActivite": "4751Z",
      "libelleActivite": "Commerce de détail de textiles en magasin spécialisé"
    },
    {
      "codeActivite": "4772A",
      "libelleActivite": "Commerce de détail de la chaussure"
    },
    {
      "codeActivite": "4753Z",
      "libelleActivite": "Commerce de détail de tapis, moquettes et revêtements de murs et de sols en magasin spécialisé"
    },
    {
      "codeActivite": "4754Z",
      "libelleActivite": "Commerce de détail d''appareils électroménagers en magasin spécialisé"
    },
    {
      "codeActivite": "4759B",
      "libelleActivite": "Commerce de détail d''autres équipements du foyer"
    },
    {
      "codeActivite": "4763Z",
      "libelleActivite": "Commerce de détail d''enregistrements musicaux et vidéo en magasin spécialisé"
    },
    {
      "codeActivite": "4765Z",
      "libelleActivite": "Commerce de détail de jeux et jouets en magasin spécialisé"
    },
    {
      "codeActivite": "4775Z",
      "libelleActivite": "Commerce de détail de parfumerie et de produits de beauté en magasin spécialisé"
    },
    {
      "codeActivite": "4776Z",
      "libelleActivite": "Commerce au détail de fleurs/herboristeries"
    },
    {
      "codeActivite": "4778C",
      "libelleActivite": "Autres commerces de détail spécialisés divers"
    },
    {
      "codeActivite": "7722Z",
      "libelleActivite": "Location de vidéocassettes et disques vidéo"
    },
    {
      "codeActivite": "8553Z",
      "libelleActivite": "Enseignement de la conduite"
    },
    {
      "codeActivite": "9101Z",
      "libelleActivite": "Gestion des bibliothèques & (et) des archives"
    },
    {
      "codeActivite": "9602B",
      "libelleActivite": "Soins de beauté"
    },
    {
      "codeActivite": "4519Z",
      "libelleActivite": "Commerce d''autres véhicules automobiles"
    },
    {
      "codeActivite": "4719B",
      "libelleActivite": "Autres commerces de détail en magasin non spécialisé"
    },
    {
      "codeActivite": "4771Z",
      "libelleActivite": "Commerce de détail d''habillement en magasin spécialisé"
    },
    {
      "codeActivite": "4782Z",
      "libelleActivite": "Commerce de détail de textiles, d''habillement et de chaussures sur éventaires et marchés"
    },
    {
      "codeActivite": "4759A",
      "libelleActivite": "Commerce de détail de meubles"
    },
    {
      "codeActivite": "4761Z",
      "libelleActivite": "Commerce de détail de livres en magasin spécialisé"
    },
    {
      "codeActivite": "4764Z",
      "libelleActivite": "Commerce de détail d''articles de sport en magasin spécialisé"
    },
    {
      "codeActivite": "4772B",
      "libelleActivite": "Commerce de détail de maroquinerie et d''articles de voyage"
    },
    {
      "codeActivite": "4777Z",
      "libelleActivite": "Commerce de détail d''articles d''horlogerie et de bijouterie en magasin spécialisé"
    },
    {
      "codeActivite": "4778B",
      "libelleActivite": "Commerces de détail de charbons et combustibles"
    },
    {
      "codeActivite": "4779Z",
      "libelleActivite": "Commerce de détail de biens d''occasion en magasin"
    },
    {
      "codeActivite": "7729Z",
      "libelleActivite": "Location et location-bail d''autres biens personnels et domestiques"
    },
    {
      "codeActivite": "8891A",
      "libelleActivite": "Accueil de jeunes enfants"
    },
    {
      "codeActivite": "9602A",
      "libelleActivite": "Coiffure"
    },
    {
      "codeActivite": "1814Z",
      "libelleActivite": "Reliure et activités connexes"
    },
    {
      "codeActivite": "3220Z",
      "libelleActivite": "Fabrication d''instruments de musique"
    }
  ]
}'::jsonb as js),
v as (select (jsonb_each(js)).*
from js) 
select key as secteur_covid,
jsonb_array_elements(v.value)->>'codeActivite' as code_activite, 
jsonb_array_elements(v.value)->>'libelleActivite' as libelle_activite
from v;

-- create index idx_etablissement_periode_urssaf_siret_siren on etablissement_periode_urssaf (siret, siren, version)
drop materialized view v_summaries;
drop materialized view v_last_effectif;

create materialized view v_last_effectif as
with last_effectif as (select siret, siren, last(effectif order by periode) as effectif,
last(periode order by periode) as periode
  from etablissement_periode_urssaf0 where effectif is not null
  group by siret, siren)
select siret, siren, effectif, sum(coalesce(effectif,0)) OVER (PARTITION BY siren) as effectif_entreprise, periode
from last_effectif;


create unique index idx_v_last_effectif_siret
  on v_last_effectif (siret);
create index idx_v_last_effectif_effectif 
  on v_last_effectif (effectif);

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
    di.chiffre_affaire, 
	di.prev_chiffre_affaire,
	di.arrete_bilan,
	di.exercice_diane, 
	di.variation_ca, 
	di.resultat_expl, 
	di.prev_resultat_expl,
	di.excedent_brut_d_exploitation,
	di.prev_excedent_brut_d_exploitation,
	ef.effectif, ef.effectif_entreprise, ef.periode as date_effectif,
    n.libelle_n5, n.libelle_n1, et.code_activite, 
	coalesce(ep.last_procol, 'in_bonis') as last_procol,
	ep.date_effet as date_last_procol,
    coalesce(ap.ap, false) as activite_partielle,
    ap.heure_consomme as apconso_heure_consomme,
    ap.montant as apconso_montant,
    u.dette[1] > u.dette[2] or u.dette[2] > u.dette[3] as hausse_urssaf,
    u.dette[1] as dette_urssaf,
	u.periode_urssaf,
	u.presence_part_salariale,
    s.alert,
    et.siege, g.raison_sociale as raison_sociale_groupe,
    ti.code_commune is not null as territoire_industrie,
	ti.code_terrind as code_territoire_industrie,
	ti.libelle_terrind as libelle_territoire_industrie,
	cj.libelle statut_juridique_n3,
	cj2.libelle statut_juridique_n2,
	cj1.libelle statut_juridique_n1,
	et.creation as date_ouverture_etablissement,
	en.creation as date_creation_entreprise,
	aet.last_list,
	aet.last_alert,
	coalesce(sco.secteur_covid, 'nonSecteurCovid') as secteur_covid,
  et.etat_administratif,
  en.etat_administratif as etat_administratif_entreprise
  from last_liste l
  	inner join etablissement0 et on true
    inner join entreprise0 en on en.siren = et.siren
    inner join v_etablissement_raison_sociale etrs on etrs.id_etablissement = et.id
    inner join v_roles r on et.siren = r.siren
    inner join departements d on d.code = et.departement
    left join v_naf n on n.code_n5 = et.code_activite
	  left join secteurs_covid sco on sco.code_activite = et.code_activite
    left join score0 s on et.siret = s.siret and s.libelle_liste = l.last_liste
    left join v_alert_etablissement aet on aet.siret = et.siret
    left join v_alert_entreprise aen on aen.siren = et.siren
    left join v_last_effectif ef on ef.siret = et.siret
    left join v_hausse_urssaf u on u.siret = et.siret
    left join v_apdemande ap on ap.siret = et.siret
    left join v_last_procol ep on ep.siret = et.siret
    left join v_diane_variation_ca di on di.siren = et.siren
    left join entreprise_ellisphere0 g on g.siren = et.siren
    left join terrind ti on ti.code_commune = et.code_commune
	left join categorie_juridique cj on cj.code = en.statut_juridique
	left join categorie_juridique cj2 on substring(cj.code from 0 for 3) = cj2.code
	left join categorie_juridique cj1 on substring(cj.code from 0 for 2) = cj1.code;

create index idx_v_summaries_raison_sociale on v_summaries using gin (siret gin_trgm_ops, raison_sociale gin_trgm_ops);
create index idx_v_summaries_siren on v_summaries (siren);
create index idx_v_summaries_siret on v_summaries (siret);
create index idx_v_summaries_order_raison_sociale on v_summaries (raison_sociale, siret);
create index idx_v_summaries_roles on v_summaries using gin (roles);
create index idx_v_summaries_score on v_summaries (
  valeur_score desc,
  siret,
  siege,
  effectif,
  effectif_entreprise,
  code_departement,
  last_procol,
  chiffre_affaire,
  etat_administratif) where alert != 'Pas d''alerte';

drop function get_currentscore;
CREATE OR REPLACE FUNCTION public.get_currentscore(
    roles_users text[], nblimit integer, nboffset integer, libelle_liste text, siret_expression text,
    raison_sociale_expression text, ignore_roles boolean, ignore_zone boolean, username text,
    siege_uniquement boolean, order_by text, alert_only boolean, last_procol text[], departements text[],
    suivi boolean, effectif_min integer, effectif_max integer, sirens text[], activites text[],
    effectif_min_entreprise integer, effectif_max_entreprise integer, ca_min integer, ca_max integer,
    exclude_secteurs_covid text[], etat_administratif text)
  RETURNS TABLE(
    siret text, siren text, raison_sociale text, commune text, libelle_departement text, 
    code_departement text, valeur_score real, detail_score jsonb, first_alert boolean, 
    chiffre_affaire real, arrete_bilan date, exercice_diane integer, variation_ca real, 
    resultat_expl real, effectif real, effectif_entreprise real, libelle_n5 text, libelle_n1 text, code_activite text, 
    last_procol text, activite_partielle boolean, apconso_heure_consomme integer, 
    apconso_montant integer, hausse_urssaf boolean, dette_urssaf real, alert text, 
    nb_total bigint, nb_f1 bigint, nb_f2 bigint, visible boolean, in_zone boolean, 
    followed boolean, followed_enterprise boolean, siege boolean, raison_sociale_groupe text, 
    territoire_industrie boolean, comment text, category text, since timestamp without time zone, 
    urssaf boolean, dgefp boolean, score boolean, bdf boolean, secteur_covid text, excedent_brut_d_exploitation real,
    etat_administratif text, etat_administratif_entreprise text) 
  LANGUAGE 'sql'
  COST 100
  IMMUTABLE PARALLEL UNSAFE
  ROWS 1000
AS $BODY$
select 
    s.siret, s.siren, s.raison_sociale, s.commune,
    s.libelle_departement, s.code_departement,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.valeur_score end as valeur_score,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.detail_score end as detail_score,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.first_alert end as first_alert,
    s.chiffre_affaire, s.arrete_bilan, s.exercice_diane, s.variation_ca, s.resultat_expl, s.effectif, s.effectif_entreprise,
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
    (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).bdf,
    s.secteur_covid, s.excedent_brut_d_exploitation, s.etat_administratif, s.etat_administratif_entreprise
  from v_summaries s
    left join v_naf n on n.code_n5 = s.code_activite
    left join etablissement_follow f on f.active and f.siret = s.siret and f.username = $9
    left join v_entreprise_follow fe on fe.siren = s.siren and fe.username = $9
    where   
    (s.roles && $1 or $7)
    and s.alert != 'Pas d''alerte'
	  and (s.code_departement=any($1) or $8)
    and (s.siege or not $10)
    and (s.last_procol = any($13) or $13 is null)
    and (s.code_departement=any($14) or $14 is null)
    and (s.effectif >= $16 or $16 is null)
    and (s.effectif <= $17 or $17 is null)
    and (s.effectif_entreprise >= $20 or $20 is null)
    and (s.effectif_entreprise <= $21 or $21 is null)
    and (s.chiffre_affaire >= $22 or $22 is null)
    and (s.chiffre_affaire <= $23 or $23 is null)
    and (fe.siren is not null = $15 or $15 is null)
	  and (s.raison_sociale ilike $6 or s.siret ilike $5 or coalesce($5, $6) is null)
    and 'score' = any($1)
    and (not (s.secteur_covid = any($24)) or $24 is null)
    and (n.code_n1 = any($19) or $19 is null)
    and (s.etat_administratif = $25 or $25 is null)
  order by s.alert, s.valeur_score desc, s.siret
  limit $2 offset $3
$BODY$;

drop function get_score;
CREATE OR REPLACE FUNCTION public.get_score(
    roles_users text[], nblimit integer, nboffset integer, libelle_liste text, siret_expression text,
    raison_sociale_expression text, ignore_roles boolean, ignore_zone boolean, username text,
    siege_uniquement boolean, order_by text, alert_only boolean, last_procol text[], departements text[],
    suivi boolean, effectif_min integer, effectif_max integer, sirens text[], activites text[],
    effectif_min_entreprise integer, effectif_max_entreprise integer, ca_min integer, ca_max integer,
    exclude_secteurs_covid text[], etat_administratif text)
  RETURNS TABLE(
    siret text, siren text, raison_sociale text, commune text, libelle_departement text, 
    code_departement text, valeur_score real, detail_score jsonb, first_alert boolean, 
    chiffre_affaire real, arrete_bilan date, exercice_diane integer, variation_ca real, 
    resultat_expl real, effectif real, effectif_entreprise real, libelle_n5 text, libelle_n1 text, code_activite text, 
    last_procol text, activite_partielle boolean, apconso_heure_consomme integer, 
    apconso_montant integer, hausse_urssaf boolean, dette_urssaf real, alert text, 
    nb_total bigint, nb_f1 bigint, nb_f2 bigint, visible boolean, in_zone boolean, 
    followed boolean, followed_enterprise boolean, siege boolean, raison_sociale_groupe text, 
    territoire_industrie boolean, comment text, category text, since timestamp without time zone, 
    urssaf boolean, dgefp boolean, score boolean, bdf boolean, secteur_covid text, excedent_brut_d_exploitation real,
    etat_administratif text, etat_administratif_entreprise text) 
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
    s.chiffre_affaire, s.arrete_bilan, s.exercice_diane, s.variation_ca, s.resultat_expl, s.effectif, s.effectif_entreprise,
    s.libelle_n5, s.libelle_n1, s.code_activite, s.last_procol,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp then s.activite_partielle end as activite_partielle,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp then s.apconso_heure_consomme end as apconso_heure_consomme,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp then s.apconso_montant end as apconso_montant,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).urssaf then s.hausse_urssaf end as hausse_urssaf,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).urssaf then s.dette_urssaf end as dette_urssaf,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then sc.alert end,
    count(*) over () as nb_total,
    count(case when sc.alert='Alerte seuil F1' and (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then 1 end) over () as nb_f1,
    count(case when sc.alert='Alerte seuil F2' and (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then 1 end) over () as nb_f2,
    (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).visible, 
    (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).in_zone, 
	  f.id is not null as followed_etablissement,
    fe.siren is not null as followed_entreprise,
    s.siege, s.raison_sociale_groupe, territoire_industrie,
    f.comment, f.category, f.since,
    (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).urssaf,
    (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp,
    (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score,
    (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).bdf,
    s.secteur_covid, s.excedent_brut_d_exploitation, s.etat_administratif, s.etat_administratif_entreprise
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
    and (s.effectif_entreprise >= $20 or $20 is null)
    and (s.effectif_entreprise <= $21 or $21 is null)
    and (s.chiffre_affaire >= $22 or $22 is null)
    and (s.chiffre_affaire <= $23 or $23 is null)
    and (fe.siren is not null = $15 or $15 is null)
	  and (s.raison_sociale ilike $6 or s.siret ilike $5 or coalesce($5, $6) is null)
    and 'score' = any($1)
    and (not (s.secteur_covid = any($24)) or $24 is null)
    and (n.code_n1 = any($19) or $19 is null)
    and (s.etat_administratif = $25 or $25 is null)
  order by sc.alert, sc.score desc, sc.siret
  limit $2 offset $3
$BODY$;

drop function get_follow;
create or replace function get_follow (
    roles_users text[], nblimit integer, nboffset integer, libelle_liste text, siret_expression text,
    raison_sociale_expression text, ignore_roles boolean, ignore_zone boolean, username text,
    siege_uniquement boolean, order_by text, alert_only boolean, last_procol text[], departements text[],
    suivi boolean, effectif_min integer, effectif_max integer, sirens text[], activites text[],
    effectif_min_entreprise integer, effectif_max_entreprise integer, ca_min integer, ca_max integer,
    exclude_secteurs_covid text[], etat_administratif text)
  RETURNS TABLE(
    siret text, siren text, raison_sociale text, commune text, libelle_departement text, 
    code_departement text, valeur_score real, detail_score jsonb, first_alert boolean, 
    chiffre_affaire real, arrete_bilan date, exercice_diane integer, variation_ca real, 
    resultat_expl real, effectif real, effectif_entreprise real, libelle_n5 text, libelle_n1 text, code_activite text, 
    last_procol text, activite_partielle boolean, apconso_heure_consomme integer, 
    apconso_montant integer, hausse_urssaf boolean, dette_urssaf real, alert text, 
    nb_total bigint, nb_f1 bigint, nb_f2 bigint, visible boolean, in_zone boolean, 
    followed boolean, followed_enterprise boolean, siege boolean, raison_sociale_groupe text, 
    territoire_industrie boolean, comment text, category text, since timestamp without time zone, 
    urssaf boolean, dgefp boolean, score boolean, bdf boolean, secteur_covid text, excedent_brut_d_exploitation real,
    etat_administratif text, etat_administratif_entreprise text) 
  LANGUAGE 'sql'
  COST 100
  IMMUTABLE PARALLEL UNSAFE
  ROWS 1000 
  AS $BODY$
  select 
    s.siret, s.siren, s.raison_sociale, s.commune,
    s.libelle_departement, s.code_departement,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.valeur_score end as valeur_score,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.detail_score end as detail_score,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.first_alert end as first_alert,
    s.chiffre_affaire, s.arrete_bilan, s.exercice_diane, s.variation_ca, s.resultat_expl, s.effectif, s.effectif_entreprise,
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
    (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).bdf,
    s.secteur_covid, s.excedent_brut_d_exploitation, s.etat_administratif, s.etat_administratif_entreprise
  from v_summaries s
    inner join etablissement_follow f on f.active and f.siret = s.siret and f.username = $9
    inner join v_entreprise_follow fe on fe.siren = s.siren and fe.username = $9
    order by f.id
  limit $2 offset $3
$BODY$;

drop function get_search;
create or replace function get_search (
    roles_users text[], nblimit integer, nboffset integer, libelle_liste text, siret_expression text,
    raison_sociale_expression text, ignore_roles boolean, ignore_zone boolean, username text,
    siege_uniquement boolean, order_by text, alert_only boolean, last_procol text[], departements text[],
    suivi boolean, effectif_min integer, effectif_max integer, sirens text[], activites text[],
    effectif_min_entreprise integer, effectif_max_entreprise integer, ca_min integer, ca_max integer,
    exclude_secteurs_covid text[], etat_administratif text)
  RETURNS TABLE(
    siret text, siren text, raison_sociale text, commune text, libelle_departement text, 
    code_departement text, valeur_score real, detail_score jsonb, first_alert boolean, 
    chiffre_affaire real, arrete_bilan date, exercice_diane integer, variation_ca real, 
    resultat_expl real, effectif real, effectif_entreprise real, libelle_n5 text, libelle_n1 text, code_activite text, 
    last_procol text, activite_partielle boolean, apconso_heure_consomme integer, 
    apconso_montant integer, hausse_urssaf boolean, dette_urssaf real, alert text, 
    nb_total bigint, nb_f1 bigint, nb_f2 bigint, visible boolean, in_zone boolean, 
    followed boolean, followed_enterprise boolean, siege boolean, raison_sociale_groupe text, 
    territoire_industrie boolean, comment text, category text, since timestamp without time zone, 
    urssaf boolean, dgefp boolean, score boolean, bdf boolean, secteur_covid text, excedent_brut_d_exploitation real,
    etat_administratif text, etat_administratif_entreprise text) 
  as $$
  select 
    s.siret, s.siren, s.raison_sociale, s.commune,
    s.libelle_departement, s.code_departement,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.valeur_score end as valeur_score,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.detail_score end as detail_score,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.first_alert end as first_alert,
    s.chiffre_affaire, s.arrete_bilan, s.exercice_diane, s.variation_ca, s.resultat_expl, s.effectif, s.effectif_entreprise,
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
    (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).bdf,
    s.secteur_covid, s.excedent_brut_d_exploitation, s.etat_administratif, s.etat_administratif_entreprise
  from v_summaries s
    left join etablissement_follow f on f.active and f.siret = s.siret and f.username = $9
    left join v_entreprise_follow fe on fe.siren = s.siren and fe.username = $9
    left join v_naf n on n.code_n5 = s.code_activite
  where   
    (s.raison_sociale ilike $6 or s.siret ilike $5)
    and (s.roles && $1 or $7)               -- plus d'ignoreRoles mais on garde la possibilité de le réimplémenter
    and (s.code_departement=any($1) or $8)  -- idem pour ignoreZone
    and (s.code_departement=any($14) or $14 is null)
    and (s.effectif >= $16 or $16 is null)
    and (s.effectif_entreprise >= $20 or $20 is null)
    and (s.effectif_entreprise <= $21 or $21 is null)
    and (s.chiffre_affaire >= $22 or $22 is null)
    and (s.chiffre_affaire <= $23 or $23 is null)
    and (n.code_n1 = any($19) or $19 is null)
    and (s.siege or not $10)
    and (not (s.secteur_covid = any($24)) or $24 is null)
    and (s.etat_administratif = $25 or $25 is null)
  order by s.raison_sociale, s.siret
  limit $2 offset $3
$$ language sql immutable;

drop function get_brother;
create or replace function get_brother (
    roles_users text[], nblimit integer, nboffset integer, libelle_liste text, siret_expression text,
    raison_sociale_expression text, ignore_roles boolean, ignore_zone boolean, username text,
    siege_uniquement boolean, order_by text, alert_only boolean, last_procol text[], departements text[],
    suivi boolean, effectif_min integer, effectif_max integer, sirens text[], activites text[],
    effectif_min_entreprise integer, effectif_max_entreprise integer, ca_min integer, ca_max integer,
    exclude_secteurs_covid text[], etat_administratif text)
  RETURNS TABLE(
    siret text, siren text, raison_sociale text, commune text, libelle_departement text, 
    code_departement text, valeur_score real, detail_score jsonb, first_alert boolean, 
    chiffre_affaire real, arrete_bilan date, exercice_diane integer, variation_ca real, 
    resultat_expl real, effectif real, effectif_entreprise real, libelle_n5 text, libelle_n1 text, code_activite text, 
    last_procol text, activite_partielle boolean, apconso_heure_consomme integer, 
    apconso_montant integer, hausse_urssaf boolean, dette_urssaf real, alert text, 
    nb_total bigint, nb_f1 bigint, nb_f2 bigint, visible boolean, in_zone boolean, 
    followed boolean, followed_enterprise boolean, siege boolean, raison_sociale_groupe text, 
    territoire_industrie boolean, comment text, category text, since timestamp without time zone, 
    urssaf boolean, dgefp boolean, score boolean, bdf boolean, secteur_covid text, excedent_brut_d_exploitation real,
    etat_administratif text, etat_administratif_entreprise text) 
as $$
  select 
    s.siret, s.siren, s.raison_sociale, s.commune,
    s.libelle_departement, s.code_departement,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.valeur_score end as valeur_score,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.detail_score end as detail_score,
    case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.first_alert end as first_alert,
    s.chiffre_affaire, s.arrete_bilan, s.exercice_diane, s.variation_ca, s.resultat_expl, s.effectif, s.effectif_entreprise,
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
    (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).bdf,
    s.secteur_covid, s.excedent_brut_d_exploitation, s.etat_administratif, s.etat_administratif_entreprise
  from v_summaries s
    left join etablissement_follow f on f.active and f.siret = s.siret and f.username = $9
    left join v_entreprise_follow fe on fe.siren = s.siren and fe.username = $9
  where  s.siren = any($18)
  order by s.effectif desc, s.siret
  limit $2 offset $3
$$ language sql immutable;