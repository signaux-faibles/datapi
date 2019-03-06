package dalib

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"time"

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

	if bcrypt.CompareHashAndPassword(hash, []byte(password)) == nil {
		return User{
			Email: email,
			Scope: ToScope(user.Value["scope"].([]interface{})),
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

// BucketPolicies is a slice of policies with apply method
type BucketPolicies []BucketPolicy

// CurrentBucketPolicies is a cache object
var CurrentBucketPolicies BucketPolicies

// LoadPolicies returns bucket policies at a given time
func LoadPolicies(date *time.Time) BucketPolicies {
	var params QueryParams
	params.Key = Map{
		"type": "policy",
	}
	params.Date = date

	policies, err := Query("auth", params, Tags{"auth", "admin"})
	if err != nil {
		fmt.Println(err)
	}

	var bucketPolicies BucketPolicies

	for _, bp := range policies {
		match := bp.Key["match"]
		read := ToScope(bp.Value["read"].([]interface{}))
		write := ToScope(bp.Value["write"].([]interface{}))
		key, err := ToMap(bp.Value["key"])
		if err != nil {
			fmt.Println(err)
		}

		matchre := regexp.MustCompilePOSIX(match)

		bucketPolicies = append(bucketPolicies, BucketPolicy{
			Name:  bp.Key["name"],
			Match: matchre.MatchString,
			Read:  read,
			Key:   key,
			Write: write,
		})

		sort.Slice(bucketPolicies, func(i, j int) bool { return bucketPolicies[i].Name < bucketPolicies[j].Name })
	}

	return BucketPolicies(bucketPolicies)
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
