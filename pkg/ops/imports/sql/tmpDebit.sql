drop table if exists tmp_debit;
create table tmp_debit
(
    siret text,
    numéro_compte text,
    numéro_écart_négatif text,
    date_traitement date,
    part_ouvrière real,
    part_patronale real,
    numéro_historique_écart_négatif int,
    état_compte int,
    code_procédure_collective int,
    période_start date,
    période_end date,
    code_opération_écart_négatif int,
    code_motif_écart_négatif int,
    recours boolean
);