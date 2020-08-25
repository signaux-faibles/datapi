create table regions (
  id integer primary key,
  code text,
  libelle text,
);

INSERT INTO `regions` 
  (id, code, libelle)
VALUES 
  (1,'01','Guadeloupe'),
  (2,'02','Martinique'),
  (3,'03','Guyane'),
  (4,'04','La Réunion'),
  (5,'06','Mayotte'),
  (6,'11','Île-de-France'),
  (7,'24','Centre-Val de Loire'),
  (8,'27','Bourgogne-Franche-Comté'),
  (9,'28','Normandie'),
  (10,'32','Hauts-de-France'),
  (11,'44','Grand Est'),
  (12,'52','Pays de la Loire'),
  (13,'53','Bretagne'),
  (14,'75','Nouvelle-Aquitaine'),
  (15,'76','Occitanie'),
  (16,'84','Auvergne-Rhône-Alpes'),
  (17,'93','Provence-Alpes-Côte d''Azur'),
  (18,'94','Corse'),
  (19,'COM','Collectivités d''Outre-Mer');

create table departements (
  code text primary key,
  id_region integer references regions (id),
  libelle text
);

insert into departements
select d.code, r.id, d.libelle from
(select '01' as code, 'Ain' as libelle, '84' as code_region
union select '02' as code, 'Aisne' as libelle, '32' as code_region
union select '03' as code, 'Allier' as libelle, '84' as code_region
union select '04' as code, 'Alpes-de-Haute-Provence' as libelle, '93' as code_region
union select '05' as code, 'Hautes-Alpes' as libelle, '93' as code_region
union select '06' as code, 'Alpes-Maritimes' as libelle, '93' as code_region
union select '07' as code, 'Ardèche' as libelle, '84' as code_region
union select '08' as code, 'Ardennes' as libelle, '44' as code_region
union select '09' as code, 'Ariège' as libelle, '76' as code_region
union select '10' as code, 'Aube' as libelle, '44' as code_region
union select '11' as code, 'Aude' as libelle, '76' as code_region
union select '12' as code, 'Aveyron' as libelle, '76' as code_region
union select '13' as code, 'Bouches-du-Rhône' as libelle, '93' as code_region
union select '14' as code, 'Calvados' as libelle, '28' as code_region
union select '15' as code, 'Cantal' as libelle, '84' as code_region
union select '16' as code, 'Charente' as libelle, '75' as code_region
union select '17' as code, 'Charente-Maritime' as libelle, '75' as code_region
union select '18' as code, 'Cher' as libelle, '24' as code_region
union select '19' as code, 'Corrèze' as libelle, '75' as code_region
union select '21' as code, 'Côte-d''Or' as libelle, '27' as code_region
union select '22' as code, 'Côtes-d''Armor' as libelle, '53' as code_region
union select '23' as code, 'Creuse' as libelle, '75' as code_region
union select '24' as code, 'Dordogne' as libelle, '75' as code_region
union select '25' as code, 'Doubs' as libelle, '27' as code_region
union select '26' as code, 'Drôme' as libelle, '84' as code_region
union select '27' as code, 'Eure' as libelle, '28' as code_region
union select '28' as code, 'Eure-et-Loir' as libelle, '24' as code_region
union select '29' as code, 'Finistère' as libelle, '53' as code_region
union select '2A' as code, 'Corse-du-Sud' as libelle, '94' as code_region
union select '2B' as code, 'Haute-Corse' as libelle, '94' as code_region
union select '30' as code, 'Gard' as libelle, '76' as code_region
union select '31' as code, 'Haute-Garonne' as libelle, '76' as code_region
union select '32' as code, 'Gers' as libelle, '76' as code_region
union select '33' as code, 'Gironde' as libelle, '75' as code_region
union select '34' as code, 'Hérault' as libelle, '76' as code_region
union select '35' as code, 'Ille-et-Vilaine' as libelle, '53' as code_region
union select '36' as code, 'Indre' as libelle, '24' as code_region
union select '37' as code, 'Indre-et-Loire' as libelle, '24' as code_region
union select '38' as code, 'Isère' as libelle, '84' as code_region
union select '39' as code, 'Jura' as libelle, '27' as code_region
union select '40' as code, 'Landes' as libelle, '75' as code_region
union select '41' as code, 'Loir-et-Cher' as libelle, '24' as code_region
union select '42' as code, 'Loire' as libelle, '84' as code_region
union select '43' as code, 'Haute-Loire' as libelle, '84' as code_region
union select '44' as code, 'Loire-Atlantique' as libelle, '52' as code_region
union select '45' as code, 'Loiret' as libelle, '24' as code_region
union select '46' as code, 'Lot' as libelle, '76' as code_region
union select '47' as code, 'Lot-et-Garonne' as libelle, '75' as code_region
union select '48' as code, 'Lozère' as libelle, '76' as code_region
union select '49' as code, 'Maine-et-Loire' as libelle, '52' as code_region
union select '50' as code, 'Manche' as libelle, '28' as code_region
union select '51' as code, 'Marne' as libelle, '44' as code_region
union select '52' as code, 'Haute-Marne' as libelle, '44' as code_region
union select '53' as code, 'Mayenne' as libelle, '52' as code_region
union select '54' as code, 'Meurthe-et-Moselle' as libelle, '44' as code_region
union select '55' as code, 'Meuse' as libelle, '44' as code_region
union select '56' as code, 'Morbihan' as libelle, '53' as code_region
union select '57' as code, 'Moselle' as libelle, '44' as code_region
union select '58' as code, 'Nièvre' as libelle, '27' as code_region
union select '59' as code, 'Nord' as libelle, '32' as code_region
union select '60' as code, 'Oise' as libelle, '32' as code_region
union select '61' as code, 'Orne' as libelle, '28' as code_region
union select '62' as code, 'Pas-de-Calais' as libelle, '32' as code_region
union select '63' as code, 'Puy-de-Dôme' as libelle, '84' as code_region
union select '64' as code, 'Pyrénées-Atlantiques' as libelle, '75' as code_region
union select '65' as code, 'Hautes-Pyrénées' as libelle, '76' as code_region
union select '66' as code, 'Pyrénées-Orientales' as libelle, '76' as code_region
union select '67' as code, 'Bas-Rhin' as libelle, '44' as code_region
union select '68' as code, 'Haut-Rhin' as libelle, '44' as code_region
union select '69' as code, 'Rhône' as libelle, '84' as code_region
union select '70' as code, 'Haute-Saône' as libelle, '27' as code_region
union select '71' as code, 'Saône-et-Loire' as libelle, '27' as code_region
union select '72' as code, 'Sarthe' as libelle, '52' as code_region
union select '73' as code, 'Savoie' as libelle, '84' as code_region
union select '74' as code, 'Haute-Savoie' as libelle, '84' as code_region
union select '75' as code, 'Paris' as libelle, '11' as code_region
union select '76' as code, 'Seine-Maritime' as libelle, '28' as code_region
union select '77' as code, 'Seine-et-Marne' as libelle, '11' as code_region
union select '78' as code, 'Yvelines' as libelle, '11' as code_region
union select '79' as code, 'Deux-Sèvres' as libelle, '75' as code_region
union select '80' as code, 'Somme' as libelle, '32' as code_region
union select '81' as code, 'Tarn' as libelle, '76' as code_region
union select '82' as code, 'Tarn-et-Garonne' as libelle, '76' as code_region
union select '83' as code, 'Var' as libelle, '93' as code_region
union select '84' as code, 'Vaucluse' as libelle, '93' as code_region
union select '85' as code, 'Vendée' as libelle, '52' as code_region
union select '86' as code, 'Vienne' as libelle, '75' as code_region
union select '87' as code, 'Haute-Vienne' as libelle, '75' as code_region
union select '88' as code, 'Vosges' as libelle, '44' as code_region
union select '89' as code, 'Yonne' as libelle, '27' as code_region
union select '90' as code, 'Territoire de Belfort' as libelle, '27' as code_region
union select '91' as code, 'Essonne' as libelle, '11' as code_region
union select '92' as code, 'Hauts-de-Seine' as libelle, '11' as code_region
union select '93' as code, 'Seine-Saint-Denis' as libelle, '11' as code_region
union select '94' as code, 'Val-de-Marne' as libelle, '11' as code_region
union select '95' as code, 'Val-d''Oise' as libelle, '11' as code_region
union select '971' as code, 'Guadeloupe' as libelle, '01' as code_region
union select '972' as code, 'Martinique' as libelle, '02' as code_region
union select '973' as code, 'Guyane' as libelle, '03' as code_region
union select '974' as code, 'La Réunion' as libelle, '04' as code_region
union select '976' as code, 'Mayotte' as libelle, '06' as code_region) d
inner join regions r on r.code = d.code_region;


create table naf (
  id integer,
  niveau integer,
  code text,
  libelle text,
  id_parent integer,
  id_n1 integer
);

-- niveau 1
insert into naf (id, id_n1, niveau, code, libelle) values
(0,0,1,'A','Agriculture, sylviculture et pêche'),
(1,1,1,'B','Industries extractives'),
(2,2,1,'C','Industrie manufacturière'),
(3,3,1,'D','Production et distribution d''électricité, de gaz, de vapeur et d''air conditionné'),
(4,4,1,'E','Production et distribution d''eau ; assainissement, gestion des déchets et dépollution'),
(5,5,1,'F','Construction'),
(6,6,1,'G','Commerce ; réparation d''automobiles et de motocycles'),
(7,7,1,'H','Transports et entreposage'),
(8,8,1,'I','Hébergement et restauration'),
(9,9,1,'J','Information et communication'),
(10,10,1,'K','Activités financières et d''assurance'),
(11,11,1,'L','Activités immobilières'),
(12,12,1,'M','Activités spécialisées, scientifiques et techniques'),
(13,13,1,'N','Activités de services administratifs et de soutien'),
(14,14,1,'O','Administration publique'),
(15,15,1,'P','Enseignement'),
(16,16,1,'Q','Santé humaine et action sociale'),
(17,17,1,'R','Arts, spectacles et activités récréatives'),
(18,18,1,'S','Autres activités de services'),
(19,19,1,'T','Activités des ménages en tant qu''employeurs ; activités indifférenciées des ménages en tant que producteurs de biens et services pour usage propre'),
(20,20,1,'U','Activités extra-territoriales');

-- niveau 2
insert into naf (id, niveau, code, libelle, id_parent, id_n1) values
(20, 2,'01','Culture et production animale, chasse et services annexes',0,0),
(21, 2,'02','Sylviculture et exploitation forestière',0,0),
(22, 2,'03','Pêche et aquaculture',0,0),
(23, 2,'05','Extraction de houille et de lignite',1,1),
(24, 2,'06','Extraction d''hydrocarbures',1,1),
(25, 2,'07','Extraction de minerais métalliques',1,1),
(26, 2,'08','Autres industries extractives',1,1),
(27, 2,'09','Services de soutien aux industries extractives',1,1),
(28, 2,'10','Industries alimentaires',2,2),
(29, 2,'11','Fabrication de boissons',2,2),
(30, 2,'12','Fabrication de produits à base de tabac',2,2),
(31, 2,'13','Fabrication de textiles',2,2),
(32, 2,'14','Industrie de l''habillement',2,2),
(33, 2,'15','Industrie du cuir et de la chaussure',2,2),
(34, 2,'16','Travail du bois et fabrication d''articles en bois et en liège, à l’exception des meubles',2,2),
(35, 2,'17','Industrie du papier et du carton',2,2),
(36, 2,'18','Imprimerie et reproduction d''enregistrements',2,2),
(37, 2,'19','Cokéfaction et raffinage',2,2),
(38, 2,'20','Industrie chimique',2,2),
(39, 2,'21','Industrie pharmaceutique',2,2),
(40, 2,'22','Fabrication de produits en caoutchouc et en plastique',2,2),
(41, 2,'23','Fabrication d''autres produits minéraux non métalliques',2,2),
(42, 2,'24','Métallurgie',2,2),
(43, 2,'25','Fabrication de produits métalliques, à l’exception des machines et des équipements',2,2),
(44, 2,'26','Fabrication de produits informatiques, électroniques et optiques',2,2),
(45, 2,'27','Fabrication d''équipements électriques',2,2),
(46, 2,'28','Fabrication de machines et équipements n.c.a.',2,2),
(47, 2,'29','Industrie automobile',2,2),
(48, 2,'30','Fabrication d''autres matériels de transport',2,2),
(49, 2,'31','Fabrication de meubles',2,2),
(50, 2,'32','Autres industries manufacturières',2,2),
(51, 2,'33','Réparation et installation de machines et d''équipements ',2,2),
(52, 2,'35','Production et distribution d''électricité, de gaz, de vapeur et d''air conditionné',3,3),
(53, 2,'36','Captage, traitement et distribution d''eau ',4,4),
(54, 2,'37','Collecte et traitement des eaux usées',4,4),
(55, 2,'38','Collecte, traitement et élimination des déchets',4,4),
(56, 2,'39','Dépollution et autres services de gestion des déchets',4,4),
(57, 2,'41','Construction de bâtiments ',5,5),
(58, 2,'42','Génie civil',5,5),
(59, 2,'43','Travaux de construction spécialisés ',5,5),
(60, 2,'45','Commerce et réparation d''automobiles et de motocycles',6,6),
(61, 2,'46','Commerce de gros, à l’exception des automobiles et des motocycles',6,6),
(62, 2,'47','Commerce de détail, à l’exception des automobiles et des motocycles',6,6),
(63, 2,'49','Transports terrestres et transport par conduites',7,7),
(64, 2,'50','Transports par eau',7,7),
(65, 2,'51','Transports aériens',7,7),
(66, 2,'52','Entreposage et services auxiliaires des transports',7,7),
(67, 2,'53','Activités de poste et de courrier',7,7),
(68, 2,'55','Hébergement',8,8),
(69, 2,'56','Restauration',8,8),
(70, 2,'58','Édition',9,9),
(71, 2,'59','Production de films cinématographiques, de vidéo et de programmes de télévision',9,9),
(72, 2,'60','Programmation et diffusion',9,9),
(73, 2,'61','Télécommunications',9,9),
(74, 2,'62','Programmation, conseil et autres activités informatiques ',9,9),
(75, 2,'63','Services d''information',9,9),
(76, 2,'64','Activités des services financiers, hors assurance et caisses de retraite',10,10),
(77, 2,'65','Assurance',10,10),
(78, 2,'66','Activités auxiliaires de services financiers et d''assurance ',10,10),
(79, 2,'68','Activités immobilières',11,11),
(80, 2,'69','Activités juridiques et comptables',12,12),
(81, 2,'70','Activités des sièges sociaux',12,12),
(82, 2,'71','Activités d''architecture et d''ingénierie',12,12),
(83, 2,'72','Recherche-développement scientifique',12,12),
(84, 2,'73','Publicité et études de marché',12,12),
(85, 2,'74','Autres activités spécialisées, scientifiques et techniques',12,12),
(86, 2,'75','Activités vétérinaires',12,12),
(87, 2,'77','Activités de location et location-bail',13,13),
(88, 2,'78','Activités liées à l''emploi',13,13),
(89, 2,'79','Activités des agences de voyage, voyagistes, services de réservation et activités connexes',13,13),
(90, 2,'80','Enquêtes et sécurité',13,13),
(91, 2,'81','Services relatifs aux bâtiments et aménagement paysager',13,13),
(92, 2,'82','Activités administratives et autres activités de soutien aux entreprises',13,13),
(93, 2,'84','Administration publique et défense',14,14),
(94, 2,'85','Enseignement',15,15),
(95, 2,'86','Activités pour la santé humaine',16,16),
(96, 2,'87','Hébergement médico-social et social',16,16),
(97, 2,'88','Action sociale sans hébergement',16,16),
(98, 2,'90','Activités créatives, artistiques et de spectacle ',17,17),
(99, 2,'91','Bibliothèques, archives, musées et autres activités culturelles',17,17),
(100, 2,'92','Organisation de jeux de hasard et d''argent',17,17),
(101, 2,'93','Activités sportives, récréatives et de loisirs',17,17),
(102, 2,'94','Activités des organisations associatives',18,18),
(103, 2,'95','Réparation d''ordinateurs et de biens personnels et domestiques',18,18),
(104, 2,'96','Autres services personnels',18,18),
(105, 2,'97','Activités des ménages en tant qu''employeurs de personnel domestique',19,19),
(106, 2,'98','Activités indifférenciées des ménages en tant que producteurs de biens et services pour usage propre',19,19),
(107, 2,'99','Activités des organisations et organismes extraterritoriaux',20,20);
