package dalib

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// User describes a user with minimal informations for auth.
type User struct {
	Email string
	Scope Tags
}

// Login tests credentials and returns a User object
// Password is stored with bcrypt method
func Login(email string, password string) (User, error) {
	var key Map
	key.Set("type", "user")
	key.Set("email", email)
	o, err := Query(key, Tags{"admin"})
	if err != nil {
		return User{}, err
	}

	hash := []byte(o[0].Value["password"].(string))
	fmt.Println(hash)
	fmt.Println(bcrypt.CompareHashAndPassword(hash, []byte(password)))

	var u User
	u.Email, _ = o[0].Key.Get("email")
	u.Scope = o[0].Scope
	fmt.Println(u.Scope[0])
	return User{}, nil
}

// Load loads user object
func Load(email string) (User, error) {
	var key Map
	key.Set("type", "user")
	key.Set("email", email)
	o, err := Query(key, Tags{"admin"})
	if err != nil {
		return User{}, err
	}

	var u User
	u.Email, _ = o[0].Key.Get("email")
	u.Scope = o[0].Scope
	return User{}, nil
}
