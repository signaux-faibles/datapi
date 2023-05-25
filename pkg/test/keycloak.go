// Package test comporte le code pour simplifier les tests
package test

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"net/http"
	"net/url"
	"os"
)

var kcBearer string

// GetBearerForUser cette méthode permet d'authentifier un utilisateur et de renvoyer son bearer
// ATTENTION cette méthode est utilisée dans les tests et du coup le temps peut-etre faussé
// ainsi on aura une erreur de certificat
func GetBearerForUser(host, realm, username, password string) (string, error) {
	authEndpoint := fmt.Sprint(host, "/auth/realms/", realm, "/protocol/openid-connect/token")
	authParams := url.Values{
		"grant_type": []string{"password"},
		"client_id":  []string{viper.GetString("keycloakClient")},
		"username":   []string{username},
		"password":   []string{password},
	}
	resp, err := http.PostForm(authEndpoint, authParams)
	if err != nil {
		return "", fmt.Errorf("erreur d'authentification : %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"erreur, le statut de la réponse n'est pas correct (devrait être %d) : %d",
			http.StatusOK,
			resp.StatusCode,
		)
	}
	body, err := GetBody(resp)
	if err != nil {
		return "", fmt.Errorf("erreur pendant la lecture du corps de la réponse : %s", err)
	}
	tokens := map[string]interface{}{}
	if err = json.Unmarshal(body, &tokens); err != nil {
		return "", fmt.Errorf("erreur pendant le décodage de la réponse : %s", err)
	}
	bearer := "Bearer " + tokens["access_token"].(string)
	return bearer, nil
}

// Authenticate permet de récupérer un Bearer dans Keycloak (si activé) à partir des
// credentials stoqués dans les variables d'env KC_USERNAME et KC_PASSWORD
func Authenticate() error {
	var err error
	if !viper.GetBool("enableKeycloak") {
		kcBearer = "false keycloak"
		return nil
	}
	keycloak := viper.GetString("keycloakHostname")
	realm := viper.GetString("keycloakRealm")
	username := os.Getenv("KC_USERNAME")
	if username == "" {
		return errors.New("la variable d'environnement 'KC_USERNAME' est vide")
	}
	password := os.Getenv("KC_PASSWORD")
	if password == "" {
		return errors.New("la variable d'environnement 'KC_PASSWORD' est vide")
	}
	kcBearer, err = GetBearerForUser(keycloak, realm, username, password)
	return err
}
