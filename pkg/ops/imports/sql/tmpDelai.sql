drop table if exists tmp_delai;
create table tmp_delai
(
    siret text,
    numéro_compte text,
    numéro_contentieux text,
    date_création date,
    date_échéance date,
    durée_délai int,
    dénomination text,
    indic_6mois boolean,
    année_création int,
    montant_échéancier real,
    stade text,
    action text
);