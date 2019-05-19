package dalib

// User describes a user with minimal informations for auth.
type User struct {
	Email     string
	Scope     Tags
	Name      string
	FirstName string
	Value     map[string]interface{}
}

// // Login tests credentials and returns a User object
// // Password is stored with bcrypt method
// func Login(email string, password string) (User, error) {
// 	var params QueryParams
// 	params.Key = make(Map)
// 	params.Key["email"] = email
// 	params.Key["type"] = "credentials"

// 	data, err := Query("system", params, nil, false, nil)
// 	if err != nil {
// 		return User{}, err
// 	}

// 	if len(data) != 1 {
// 		return User{}, errors.New("invalid email or password")
// 	}

// 	user := data[0]
// 	hash := []byte(user.Value["password"].(string))
// 	scope, err := ToScope(user.Value["scope"])
// 	scope = append(scope, user.Key["email"])
// 	value := user.Value
// 	delete(value, "scope")
// 	delete(value, "password")

// 	if bcrypt.CompareHashAndPassword(hash, []byte(password)) == nil {
// 		return User{
// 			Email: email,
// 			Scope: scope,
// 			Value: value,
// 		}, nil
// 	}
// 	return User{}, errors.New("invalid email or password")
// }

// Contains checks if all elements of another Map are present in the Map
func (m Map) Contains(m2 Map) bool {
	for k, v := range m2 {
		if m[k] != v {
			return false
		}
	}
	return true
}
