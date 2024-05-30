[![CI](https://github.com/signaux-faibles/datapi/actions/workflows/pipeline.yml/badge.svg)](https://github.com/signaux-faibles/datapi/actions/workflows/pipeline.yml)

# datAPI - API des applications web signaux-faibles

## Installation/Prise en main
-   Cloner le repository: `git clone git@github.com:signaux-faibles/datapi.git`  
-   Compiler: `go build`  
-   Préparer une base de données vierge
-   Copier `config.toml.example` en `config.toml` et configurer les valeurs qu'il contient, notamment :
  
```toml
postgres = "host=localhost port=5432 user=postgres password=toto dbname=postgres"

[stats]
db_url = "postgres://postgres:toto@localhost:5432/datapilogs"
```
-   Exécuter le binaire


  Il est possible de lister les routes en exécutant la commande suivante (Il faut avoir buildé le binaire avec `go build`):
  ```bash
  ./datapi --list-routes=true
  ```

## Conteneur
-  `cd build-container`
-  `./build.sh`
-  le conteneur est généré sous forme d'image compressée sous le nom `datapi.tar.gz`
-  l'image est dépourvue de configuration, elle devra être fournie à l'exécution avec l'option `-v /foo/volume/config.toml:/app/config.toml`

## Tests
-  `go test -tags=integration -v` pour lancer les test d'intégration
-  `go test -tags=integration -v -overwriteGoldenFiles` en remplaçant les golden files par les réponses générées 
  par l'exécution des tests. C'est pratique dans le cas où de nouvelles features nécessitent de modifier les golden sources.
  __ATTENTION : __Ce cas de figure est rare. Il ne faut pas utiliser ce flag si l'on est pas sûr que le code est correct. 


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
