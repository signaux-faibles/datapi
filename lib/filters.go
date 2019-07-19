package dalib

func filterEffectif(object Object, operator string, value interface{}) bool {
	effectifMin, ok := value.(float64)
	if !ok {
		return false
	}
	effectifVal, ok := object.Value["dernier_effectif"]
	if !ok {
		return false
	}
	effectif, ok := effectifVal.(float64)
	if !ok {
		return false
	}
	if effectif >= effectifMin {
		return true
	}
	return false
}

func filterZone(object Object, operator string, value interface{}) bool {
	valueVal, ok := value.([]interface{})
	if !ok {
		return false
	}
	departementVal, ok := object.Value["departement"]
	if !ok {
		return false
	}
	departement, ok := departementVal.(string)
	if !ok {
		return false
	}
	for _, d := range valueVal {
		dept, ok := d.(string)
		if !ok {
			return false
		}
		if departement == dept {
			return true
		}
	}

	return false
}

func filterGreen(object Object) bool {
	alerteVal, ok := object.Value["alert"]
	if !ok {
		return false
	}
	alerte, ok := alerteVal.(string)
	if !ok {
		return false
	}
	if alerte == "Pas d'alerte" {
		return true
	}
	return false
}

func any(a []string, i string) bool {
	for _, v := range a {
		if i == v {
			return true
		}
	}
	return false
}

func filterNAF1(object Object, operator string, value interface{}) bool {
	naf1, ok := value.(string)
	if !ok {
		return false
	}
	activiteVal, ok := object.Value["activite"]
	if !ok {
		return false
	}
	activite, ok := activiteVal.(string)
	if !ok {
		return false
	}
	codes, ok := nafCodes[naf1]
	if any(codes, activite) {
		return true
	}
	return false
}

func filterCRP(object Object, operator string, value interface{}) bool {
	val, ok := value.(bool)
	if !ok {
		return false
	}
	connuVal, ok := object.Value["connu"]
	if !ok {
		return false
	}
	connu, ok := connuVal.(bool)
	if !ok {
		return false
	}

	if connu == val {
		return true
	}
	return false
}

func filterProcol(object Object, operator string, value interface{}) bool {
	if operator != "not in" {
		return false
	}
	paramVal, ok := value.([]interface{})
	if !ok {
		return false
	}
	procolVal, ok := object.Value["etat_procol"]
	if !ok {
		return false
	}
	procol, ok := procolVal.(string)
	if !ok {
		return false
	}
	for _, p := range paramVal {
		a, ok := p.(string)
		if !ok {
			return false
		}
		if a == procol {
			return false
		}
	}
	return true
}

var filters = map[string]func(object Object, operator string, value interface{}) bool{
	"effectif": filterEffectif,
	"naf1":     filterNAF1,
	"zone":     filterZone,
	"crp":      filterCRP,
	"procol":   filterProcol,
}
