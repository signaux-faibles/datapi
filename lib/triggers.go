package dalib

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// Trigger function is applied at insert to objects containing its subkey
type Trigger struct {
	Key      Map
	Function func(Object) (Object, error)
}

func cipherPassword(o Object) (Object, error) {
	email := o.Key["email"]
	if len(email) > 254 || !rxEmail.MatchString(email) {
		return Object{}, ErrInvalidEmail
	}

	password, ok := o.Value["password"].(string)
	if !ok {
		return Object{}, errors.New("couldn't apply trigger, object lost")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return Object{}, errors.New("couldn't apply trigger, object lost")
	}
	o.Value["password"] = string(hash)

	return o, nil
}

var triggers = []Trigger{
	Trigger{
		Key: Map{
			"type": "credentials",
		},
		Function: cipherPassword,
	},
}
