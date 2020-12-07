create table liste_description
(libelle_liste text primary key,
description text);

-- fix list names
update score set libelle_liste = substring(libelle_liste from 8) where libelle_liste ~ '^[0-9]{4}';
update liste set libelle = substring(libelle from 8) where libelle ~ '^[0-9]{4}';