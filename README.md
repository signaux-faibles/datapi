# datAPI - API des applications web signaux-faibles

## Gestion des droits
### zone d'attribution
Une entreprise et tous ses établissements sont visibles dans la zone d'un utilisateur dès lors qu'au moins un établissement est dans un de ses départements ou région d'attribution.

### diane / bilans / siren / effectifs / procol
L'utilisateur voit toutes les données de toutes les entreprises

### bdf
L'utilisateur doit impérativement disposer du role `bdf` (niveau A)
L'entreprise devient alors visible dès lors que
- l'entreprise figure sur au moins une liste d'alerte (niveau F1 ou F2) et se situe dans la zone d'attribution
- ou l'entreprise est suivie par l'utilisateur

### cotisations / débits / majorations / délais
L'utilisateur doit impérativement disposer du role `urssaf` (niveau A)
L'entreprise devient alors visible dès lors que, au choix
- l'entreprise figure sur au moins une liste d'alerte (niveau F1 ou F2) et se situe dans la zone d'attribution
- l'entreprise est suivie par l'utilisateur

### Demande et consommation d'activité partielle
L'utilisateur doit impérativement disposer du role `dgefp` (niveau A)
L'entreprise devient alors visible dès lors que, au choix
- l'entreprise figure sur au moins une liste d'alerte (niveau F1 ou F2) et se situe dans la zone d'attribution
- l'entreprise est suivie par l'utilisateur
