drop table if exists tmp_effectif;
create table tmp_effectif
(
    siret text,
    numéro_compte text,
    période date,
    effectif int
);