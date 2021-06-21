package main

type activiteCovid struct {
	codeActivite    string
	libelleActivite string
}

type secteursCovid struct {
	s1            []activiteCovid
	s1Possible    []activiteCovid
	s1bis         []activiteCovid
	s1bisPossible []activiteCovid
	s2            []activiteCovid
}

func (s secteursCovid) get(activite string) string {
	for _, i := range s.s1 {
		if i.codeActivite == activite {
			return "S1 : très probable"
		}
	}
	for _, i := range s.s1Possible {
		if i.codeActivite == activite {
			return "S1 : possible"
		}
	}
	for _, i := range s.s1bis {
		if i.codeActivite == activite {
			return "S1 bis : très probable"
		}
	}
	for _, i := range s.s1bisPossible {
		if i.codeActivite == activite {
			return "S1 bis : possible"
		}
	}
	for _, i := range s.s2 {
		if i.codeActivite == activite {
			return "S2 : très probable"
		}
	}
	return ""
}

var secteurCovid = secteursCovid{
	s1: []activiteCovid{
		{
			"4939C",
			"Téléphériques et remontées mécaniques",
		},
		{
			"5510Z",
			"Hôtels et hébergement similaire",
		},
		{
			"5520Z",
			"Hébergement touristique et autre hébergement de courte durée",
		},
		{
			"5530Z",
			"Terrains de camping et parcs pour caravanes ou véhicules de loisirs",
		},
		{
			"5610A",
			"Restauration traditionnelle",
		},
		{
			"5610B",
			"Cafétérias et autres libres-services",
		},
		{
			"5610C",
			"Restauration de type rapide",
		},
		{
			"5629A",
			"Services de restauration collective sous contrat, de cantines et restaurants d'entreprise",
		},
		{
			"5629B",
			"Services de restauration collective sous contrat, de cantines et restaurants d'entreprise",
		},
		{
			"5621Z",
			"Services des traiteurs",
		},
		{
			"5630Z",
			"Débits de boissons",
		},
		{
			"5914Z",
			"Projection de films cinématographiques et autres industries techniques du cinéma et de l'image animée",
		},
		{
			"5912Z",
			"Post-production de films cinématographiques, de vidéo et de programmes de télévision",
		},
		{
			"5913A",
			"Distribution de films cinématographiques",
		},
		{
			"7721Z",
			"Location et location-bail d'articles de loisirs et de sport",
		},
		{
			"7911Z",
			"Activités des agences de voyage",
		},
		{
			"7912Z",
			"Activités des voyagistes",
		},
		{
			"7990Z",
			"Autres services de réservation et activités connexes",
		},
		{
			"8230Z",
			"Organisation de foires, évènements publics ou privés, salons ou séminaires professionnels, congrès",
		},
		{
			"8551Z",
			"Enseignement de disciplines sportives et d'activités de loisirs",
		},
		{
			"9001Z",
			"Arts du spectacle vivant, cirques",
		},
		{
			"9002Z",
			"Activités de soutien au spectacle vivant",
		},
		{
			"9003A",
			"Création artistique relevant des arts plastiques",
		},
		{
			"9003A",
			"Artistes auteurs",
		},
		{
			"9003B",
			"Artistes auteurs",
		},
		{
			"9004Z",
			"Gestion de salles de spectacles et production de spectacles",
		},
		{
			"9102Z",
			"Gestion des musées",
		},
		{
			"9103Z",
			"Gestion des sites et monuments historiques et des attractions touristiques similaires",
		},
		{
			"9104Z",
			"Gestion des jardins botaniques et zoologiques et des réserves naturelles",
		},
		{
			"9311Z",
			"Gestion d'installations sportives",
		},
		{
			"9312Z",
			"Activités de clubs de sports",
		},
		{
			"9313Z",
			"Activité des centres de culture physique",
		},
		{
			"9319Z",
			"Autres activités liées au sport",
		},
		{
			"9321Z",
			"Activités des parcs d'attractions et parcs à thèmes, fêtes foraines",
		},
		{
			"9329Z",
			"Autres activités récréatives et de loisirs",
		},
		{
			"9604Z",
			"Entretien corporel",
		},
		{
			"5110Z",
			"Transport aérien de passagers",
		},
		{
			"4939A",
			"Transports routiers réguliers de voyageurs",
		},
		{
			"4939B",
			"Autres transports routiers de voyageurs",
		},
		{
			"5010Z",
			"Transport maritime et côtier de passagers",
		},
		{
			"5911A",
			"Production de films et de programmes pour la télévision",
		},
		{
			"5911B",
			"Production de films institutionnels et publicitaires",
		},
		{
			"5911C",
			"Production de films pour le cinéma",
		},
		{
			"7420Z",
			"Activités photographiques",
		},
		{
			"8552Z",
			"Enseignement culturel",
		},
		{
			"7430Z",
			"Traducteurs-interprètes",
		},
		{
			"4932Z",
			"Transports de voyageurs par taxis et véhicules de tourisme avec chauffeur",
		},
		{
			"7711A",
			"Location de courte durée de voitures et de véhicules automobiles légers",
		},
		{
			"2511Z",
			"Fabrication de structures métalliques et de parties de structures",
		},
		{
			"7312Z",
			"Régie publicitaire de médias",
		},
		{
			"0127Z",
			"Culture de plantes à boissons",
		},
		{
			"0121Z",
			"Culture de la vigne",
		},
		{
			"1101Z",
			"Production de boissons alcooliques distillées",
		},
		{
			"1102A",
			"Fabrication de vins effervescents",
		},
		{
			"1102B",
			"Vinification",
		},
		{
			"1103Z",
			"Fabrication de cidre et de vins de fruits",
		},
		{
			"1104Z",
			"Production d'autres boissons fermentées non distillées",
		},
	},
	s1Possible: []activiteCovid{
		{
			"7810Z",
			"Agences de mannequins",
		},
		{
			"6612Z",
			"Entreprises de détaxe et bureaux de change (changeurs manuels)",
		},
		{
			"4778C",
			"Galeries d'art ou magasins de souvenirs et de piété",
		},
		{
			"7990Z",
			"Guides conférenciers",
		},
		{
			"9200Z",
			"Exploitations de casinos",
		},
		{
			"4910Z",
			"Trains et chemins de fer touristiques",
		},
		{
			"5030Z",
			"Transport de passagers sur les fleuves, les canaux, les lacs, location de bateaux de plaisance",
		},
		{
			"7721Z",
			"Transport de passagers sur les fleuves, les canaux, les lacs, location de bateaux de plaisance",
		},
		{
			"9002Z",
			"Prestation et location de chapiteaux, tentes, structures, sonorisation, photographie, lumière et pyrotechnie",
		},
		{
			"5520Z",
			"Accueils collectifs de mineurs en hébergement touristique",
		},
	},
	s1bis: []activiteCovid{
		{
			"0311Z",
			"Pêche en mer",
		},
		{
			"0312Z",
			"Pêche en eau douce",
		},
		{
			"0321Z",
			"Aquaculture en mer",
		},
		{
			"0322Z",
			"Aquaculture en eau douce",
		},
		{
			"1105Z",
			"Fabrication de bière",
		},
		{
			"1106Z",
			"Fabrication de malt",
		},
		{
			"4617A",
			"Centrales d'achat alimentaires",
		},
		{
			"4631Z",
			"Commerce de gros de fruits et légumes",
		},
		{
			"0130Z",
			"Herboristerie/ horticulture/ commerce de gros de fleurs et plans",
		},
		{
			"4622Z",
			"Herboristerie/ horticulture/ commerce de gros de fleurs et plans",
		},
		{
			"4633Z",
			"Commerce de gros de produits laitiers, œufs, huiles et matières grasses comestibles",
		},
		{
			"4634Z",
			"Commerce de gros de boissons",
		},
		{
			"4638A",
			"Mareyage et commerce de gros de poissons, coquillages, crustacés",
		},
		{
			"4638B",
			"Commerce de gros alimentaire spécialisé divers",
		},
		{
			"4639A",
			"Commerce de gros de produits surgelés",
		},
		{
			"4639B",
			"Commerce de gros alimentaire",
		},
		{
			"4690Z",
			"Commerce de gros non spécialisé",
		},
		{
			"4641Z",
			"Commerce de gros de textiles",
		},
		{
			"4618Z",
			"Intermédiaires spécialisés dans le commerce d'autres produits spécifiques",
		},
		{
			"4642Z",
			"Commerce de gros d'habillement et de chaussures",
		},
		{
			"4649Z",
			"Commerce de gros d'autres biens domestiques",
		},
		{
			"4644Z",
			"Commerce de gros de vaisselle, verrerie et produits d'entretien",
		},
		{
			"4669C",
			"Commerce de gros de fournitures et équipements divers pour le commerce et les services",
		},
		{
			"9601A",
			"Blanchisserie-teinturerie de gros",
		},
		{
			"5920Z",
			"Enregistrement sonore et édition musicale",
		},
		{
			"5811Z",
			"Editeurs de livres (Édition de livres)",
		},
		{
			"5223Z",
			"Services auxiliaires des transports aériens",
		},
		{
			"5222Z",
			"Services auxiliaires de transport par eau",
		},
		{
			"8010Z",
			"Activités de sécurité privée",
		},
		{
			"8121Z",
			"Nettoyage courant des bâtiments",
		},
		{
			"8122Z",
			"Autres activités de nettoyage des bâtiments et nettoyage industriel",
		},
		{
			"1013B",
			"Préparation à caractère artisanal de produits de charcuterie",
		},
		{
			"1071D",
			"Pâtisserie",
		},
		{
			"4722Z",
			"Commerce de détail de viandes et de produits à base de viande en magasin spécialisé",
		},
		{
			"1412Z",
			"Fabrication de vêtements de travail",
		},
		{
			"1820Z",
			"Reproduction d'enregistrements",
		},
		{
			"2313Z",
			"Fabrication de verre creux",
		},
		{
			"2341Z",
			"Fabrication d'articles céramiques à usage domestique ou ornemental",
		},
		{
			"2571Z",
			"Fabrication de coutellerie",
		},
		{
			"2599A",
			"Fabrication d'articles métalliques ménagers",
		},
		{
			"2752Z",
			"Fabrication d'appareils ménagers non électriques",
		},
		{
			"2740Z",
			"Fabrication d'appareils d'éclairage électrique",
		},
		{
			"4321A",
			"Travaux d'installation électrique dans tous locaux",
		},
		{
			"4332C",
			"Aménagement (Agencement) de lieux de vente",
		},
		{
			"7021Z",
			"Conseil en relations publiques et communication",
		},
		{
			"7311Z",
			"Activités des agences de publicité",
		},
		{
			"7410Z",
			"Activités spécialisées de design",
		},
		{
			"7490B",
			"Activités spécialisées, scientifiques et techniques diverses",
		},
		{
			"9003B",
			"Autre création artistique",
		},
		{
			"9601B",
			"Blanchisserie-teinturerie de détail",
		},
		{
			"4799B",
			"Vente par automate",
		},
		{
			"4632B",
			"Commerce de gros de viandes et de produits à base de viande",
		},
		{
			"7810Z",
			"Activités des agences de placement de main-d'œuvre",
		},
		{
			"1413Z",
			"Couturiers",
		},
		{
			"9523Z",
			"Réparation de chaussures et d'articles en cuir",
		},
	},
	s1bisPossible: []activiteCovid{
		{
			"1051C",
			"Production de fromages sous appellation d'origine protégée ou indication géographique protégée",
		},
		{
			"4617B",
			"Autres intermédiaires du commerce en denrées et boissons",
		},
		{
			"4730Z",
			"Stations-service",
		},
		{
			"9200Z",
			"Paris sportifs",
		},
		{
			"5920Z",
			"Activités liées à la production de matrices sonores originales, sur bandes, cassettes, CD, la mise à disposition des enregistrements, leur promotion et leur distribution",
		},
		{
			"1013A",
			"Fabrication de foie gras",
		},
		{
			"4781Z",
			"Commerce de détail de viande, produits à base de viandes sur éventaires et marchés",
		},
		{
			"4776Z",
			"Commerce de détail de fleurs, en pot ou coupées, de compositions florales, de plantes et de graines",
		},
		{
			"4789Z",
			"Commerce de détail de livres sur éventaires et marchés",
		},
		{
			"6622Z",
			"Courtier en assurance voyage",
		},
		{
			"6820B",
			"Location et exploitation d'immeubles non résidentiels de réception",
		},
		{
			"8211Z",
			"Services administratifs d'assistance à la demande de visas",
		},
		{
			"4120A",
			"Construction de maisons mobiles pour les terrains de camping",
		},
		{
			"9609Z",
			"Garde d'animaux de compagnie avec ou sans hébergement",
		},
		{
			"1399Z",
			"Fabrication de dentelle et broderie",
		},
		{
			"4622Z",
			"Commerce de gros de vêtements de travail",
		},
		{
			"6010Z",
			"Edition et diffusion de programmes radios à audience locale, éditions de chaînes de télévision à audience locale",
		},
		{
			"6020B",
			"Edition et diffusion de programmes radios à audience locale, éditions de chaînes de télévision à audience locale",
		},
		{
			"3230Z",
			"Fabrication de skis, fixations et bâtons pour skis, chaussures de ski",
		},
		{
			"2591Z",
			"Fabrication de bidons de bière métalliques, tonnelets de bière métalliques, fûts de bière métalliques",
		},
	},
	s2: []activiteCovid{
		{
			"4511Z",
			"Commerce de voitures et de véhicules automobiles légers",
		},
		{
			"4719A",
			"Grands magasins",
		},
		{
			"4751Z",
			"Commerce de détail de textiles en magasin spécialisé",
		},
		{
			"4772A",
			"Commerce de détail de la chaussure",
		},
		{
			"4753Z",
			"Commerce de détail de tapis, moquettes et revêtements de murs et de sols en magasin spécialisé",
		},
		{
			"4754Z",
			"Commerce de détail d'appareils électroménagers en magasin spécialisé",
		},
		{
			"4759B",
			"Commerce de détail d'autres équipements du foyer",
		},
		{
			"4763Z",
			"Commerce de détail d'enregistrements musicaux et vidéo en magasin spécialisé",
		},
		{
			"4765Z",
			"Commerce de détail de jeux et jouets en magasin spécialisé",
		},
		{
			"4775Z",
			"Commerce de détail de parfumerie et de produits de beauté en magasin spécialisé",
		},
		{
			"4776Z",
			"Commerce au détail de fleurs/herboristeries",
		},
		{
			"4778C",
			"Autres commerces de détail spécialisés divers",
		},
		{
			"7722Z",
			"Location de vidéocassettes et disques vidéo",
		},
		{
			"8553Z",
			"Enseignement de la conduite",
		},
		{
			"9101Z",
			"Gestion des bibliothèques & (et) des archives",
		},
		{
			"9602B",
			"Soins de beauté",
		},
		{
			"4519Z",
			"Commerce d'autres véhicules automobiles",
		},
		{
			"4719B",
			"Autres commerces de détail en magasin non spécialisé",
		},
		{
			"4771Z",
			"Commerce de détail d'habillement en magasin spécialisé",
		},
		{
			"4782Z",
			"Commerce de détail de textiles, d'habillement et de chaussures sur éventaires et marchés",
		},
		{
			"4759A",
			"Commerce de détail de meubles",
		},
		{
			"4761Z",
			"Commerce de détail de livres en magasin spécialisé",
		},
		{
			"4764Z",
			"Commerce de détail d'articles de sport en magasin spécialisé",
		},
		{
			"4772B",
			"Commerce de détail de maroquinerie et d'articles de voyage",
		},
		{
			"4777Z",
			"Commerce de détail d'articles d'horlogerie et de bijouterie en magasin spécialisé",
		},
		{
			"4778B",
			"Commerces de détail de charbons et combustibles",
		},
		{
			"4779Z",
			"Commerce de détail de biens d'occasion en magasin",
		},
		{
			"7729Z",
			"Location et location-bail d'autres biens personnels et domestiques",
		},
		{
			"8891A",
			"Accueil de jeunes enfants",
		},
		{
			"9602A",
			"Coiffure",
		},
		{
			"1814Z",
			"Reliure et activités connexes",
		},
		{
			"3220Z",
			"Fabrication d'instruments de musique",
		},
	},
}
