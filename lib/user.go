package dalib

import (
	"errors"
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
	var params QueryParams
	params.Key = make(Map)
	params.Key["email"] = email
	params.Key["type"] = "credentials"

	data, err := Query("auth", params, Tags{"auth"})
	if err != nil {
		return User{}, err
	}

	if len(data) != 1 {
		fmt.Println(len(data))
		return User{}, errors.New("invalid email or password")
	}

	user := data[0]
	hash := []byte(user.Value["password"].(string))
	scope, err := ToScope(user.Value["scope"].([]interface{}))

	if bcrypt.CompareHashAndPassword(hash, []byte(password)) == nil {
		return User{
			Email: email,
			Scope: scope,
		}, nil
	}
	return User{}, errors.New("invalid email or password")
}

// BucketPolicy describes generic permissions on bucket.
// A policy sets a minimal scope for all objects that fits in the key.
// This scope is automatically added to objects when they come in and out.
type BucketPolicy struct {
	Name  string
	Match func(s string) bool
	Key   Map
	Read  Tags
	Write Tags
}

func (m Map) contains(m2 Map) bool {
	for k, v := range m2 {
		if m[k] != v {
			return false
		}
	}
	return true
}

// ApplyReadPolicies provides tags implied by bucket policies given a bucket and a key
func ApplyReadPolicies(policies BucketPolicies, bucket string, key Map) Tags {
	utags := make(map[string]struct{})

	for _, p := range policies {
		if p.Match(bucket) && key.contains(p.Key) {
			for _, tag := range p.Read {
				utags[tag] = struct{}{}
			}
		}
	}

	var tags Tags
	for t := range utags {
		tags = append(tags, t)
	}

	return tags
}

// ApplyWritePolicies provides tags implied by bucket policies given a bucket and a key
func ApplyWritePolicies(policies BucketPolicies, bucket string, key Map) Tags {
	utags := make(map[string]struct{})

	for _, p := range policies {
		if p.Match(bucket) && key.contains(p.Key) {
			for _, tag := range p.Write {
				utags[tag] = struct{}{}
			}
		}
	}

	var tags Tags
	for t := range utags {
		tags = append(tags, t)
	}

	return tags
}
