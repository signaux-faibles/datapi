  # coding: utf-8
import sys
import json
from mailmerge import MailMerge

# Le template contient à ce jour les champs :
# auteur l'auteur du document
# date_edition la date d'édition du document
# confidentialite le destinataire du document
# raison_sociale la raison sociale de l'entreprise
# siret le numéro de SIRET de l'établissement
# type_etablissement le type d'établissement siège social ou établissement secondaire
# tete_de_groupe la tête de groupe si l'entreprise fait partie d'un groupe
# departement le departement de l'établissement
# commune la commune de l'établissement
# territoire_industrie le Territoire d'industrie
# secteur_activite le secteur d'activité
# activite le libellé et le code activité
# secteurs_covid appartenance aux secteurs dits COVID-19 S1, S1 bis ou S2
# statut_juridique le statut juridique comme SAS ou SARL
# date_ouverture_etablissement la date d'ouverture de l'établissement
# date_creation_entreprise la date de création de l'entreprise
# effectif le dernier effectif
# activite_partielle demande d'activité partielle sur les 12 derniers mois ou non
# dette_sociale dette sociale en hausse sur les 3 derniers mois ou non
# part_salariale dette salariale restante ou non
# annee_exercice année du dernier exercice comptable
# ca chiffre d'affaires
# ebe excédent brut d'exploitation
# rex résultat d'exploitation
# procol dernière procédure collective
# detection_sf risque identifié par l'algorithme de détection Signaux Faibles
# date_debut_suivi date de début de suivi par l'auteur
# description_wekan description dans l'outil de suivi Kanban Wekan
template = 'template.docx'

# Lecture des données JSON depuis l'entrée standard
def get_json_input_data():
	try:
		sys.stdin.reconfigure(encoding='utf-8')
		read = sys.stdin.read()
		data = json.loads(read)
		return data
	except ValueError:
		sys.stderr.write('Erreur lors de la lecture des données JSON en entrée\n')
		sys.exit(1)

# Remplissage du modèle DOCX contenant des champs de fusion (MERGEFIELD) et écriture dans la sortie standard
def fill_template_with_data(data):
	try:
		document = MailMerge(template)
		# 3 arguments possibles :
		# 1 = auteur, 2 = date_edition, 3 = confidentialite
		args = len(sys.argv)
		if args > 3:
			confidentialite = sys.argv[3]
			document.merge(confidentialite=confidentialite)
		if args > 2:
			date_edition = sys.argv[2]
			document.merge(date_edition=date_edition)
		if args > 1:
			auteur = sys.argv[1]
			document.merge(auteur=auteur)
		document.merge_templates(data, separator='page_break')
		document.write(sys.stdout.buffer)
	except ValueError:
		sys.stderr.write('Erreur lors du remplissage du modèle DOCX\n')
		sys.exit(1)

data = get_json_input_data()
fill_template_with_data(data)
sys.exit(0)
