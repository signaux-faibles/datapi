# datAPI - API des applications web signaux-faibles

## Installation/Prise en main
-   Cloner le repository: `git clone git@github.com:signaux-faibles/datapi.git`  
-   Compiler: `go build`  
-   Préparer une base de données vierge
-   Copier `config.toml.example` en `config.toml` et configurer les valeurs qu'il contient
-   Exécuter le binaire

## Conteneur
-   `cd build-container`
-   `./build.sh`
-   le conteneur est généré sous forme d'image compressée sous le nom `datapi.tar.gz`
-   l'image est dépourvue de configuration, elle devra être fournie à l'exécution avec l'option `-v /foo/volume/config.toml:/app/config.toml`

## Tests
-   `cd test/`
-   `./test.sh`

## Gestion des droits
### zone d'attribution
Une entreprise et tous ses établissements sont visibles dans la zone d'un utilisateur dès lors qu'au moins un établissement est dans un de ses départements ou région d'attribution.

### diane / bilans / siren / effectifs / procol
L'utilisateur voit toutes les données de toutes les entreprises

### bdf
L'utilisateur doit impérativement disposer du role `bdf` (niveau A)
L'entreprise devient alors visible dès lors que
-   l'entreprise figure sur au moins une liste d'alerte (niveau F1 ou F2) et se situe dans la zone d'attribution
-   ou l'entreprise est suivie par l'utilisateur

### cotisations / débits / majorations / délais
L'utilisateur doit impérativement disposer du role `urssaf` (niveau A)
L'entreprise devient alors visible dès lors que, au choix
-   l'entreprise figure sur au moins une liste d'alerte (niveau F1 ou F2) et se situe dans la zone d'attribution
-   l'entreprise est suivie par l'utilisateur

### Demande et consommation d'activité partielle
L'utilisateur doit impérativement disposer du role `dgefp` (niveau A)
L'entreprise devient alors visible dès lors que, au choix
-   l'entreprise figure sur au moins une liste d'alerte (niveau F1 ou F2) et se situe dans la zone d'attribution
-   l'entreprise est suivie par l'utilisateur
