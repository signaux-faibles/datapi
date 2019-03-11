package dalib

import (
	"database/sql"
	"errors"
	"regexp"
	"sort"
	"sync"
	"time"
)

var policiesLock sync.Mutex

// BucketPolicy describes generic permissions on bucket.
// A policy sets a minimal scope for all objects that fits in the key.
// This scope is automatically added to objects when they come in and out.
type BucketPolicy struct {
	Name    string              `json:"name"`
	Match   func(s string) bool `json:"-"`                 // apply the policy if Match(bucket's name) == true
	Key     Map                 `json:"key,omitempty"`     // policy apply only to subkey
	Scope   Tags                `json:"scope,omitempty"`   // user needs this scope to apply policy
	Read    Tags                `json:"read,omitempty"`    // user needs this scope to read objects
	Write   Tags                `json:"write,omitempty"`   // user needs this scope to write objects
	Promote Tags                `json:"promote,omitempty"` // user inherits this scope from policy
}

// BucketPolicies is a slice of policies with apply method
type BucketPolicies []BucketPolicy

// SafeRead prevents reading policies when they're not ready
func (p BucketPolicies) SafeRead() BucketPolicies {
	policiesLock.Lock()
	defer policiesLock.Unlock()
	return p
}

// SafeUpdate prevents writing policies when they're read
func (p *BucketPolicies) SafeUpdate(np BucketPolicies) {
	policiesLock.Lock()
	defer policiesLock.Unlock()
	*p = np
}

// CurrentBucketPolicies is a cache object
var CurrentBucketPolicies BucketPolicies

// BuildPolicy translates Object to BucketPolicy
func BuildPolicy(o Object) (BucketPolicy, error) {
	name, ok := o.Key["name"]
	if !ok || name == "" {
		return BucketPolicy{}, errors.New("name subkey is mandatory for policies")
	}
	if len(o.Scope) != 0 {
		return BucketPolicy{}, errors.New("rejected policy, scope must be empty")
	}

	match := o.Value["match"].(string)
	matchre := regexp.MustCompile(match)

	key, err := ToMap(o.Value["key"])
	if err != nil {
		return BucketPolicy{}, err
	}
	read, err := ToScope(o.Value["read"])
	if err != nil {
		return BucketPolicy{}, err
	}
	write, err := ToScope(o.Value["write"])
	if err != nil {
		return BucketPolicy{}, err
	}
	promote, err := ToScope(o.Value["promote"])
	if err != nil {
		return BucketPolicy{}, err
	}

	scope, err := ToScope(o.Value["scope"])
	if err != nil {
		return BucketPolicy{}, ErrInvalidPolicy
	}

	bp := BucketPolicy{
		Name:    name,
		Match:   matchre.MatchString,
		Key:     key,
		Read:    read,
		Write:   write,
		Promote: promote,
		Scope:   scope,
	}

	return bp, nil
}

// LoadPolicies returns bucket policies at a given time
func LoadPolicies(date *time.Time, tx *sql.Tx) (BucketPolicies, error) {
	var params QueryParams
	params.Key = Map{
		"type": "policy",
	}
	params.Date = date

	policies, err := Query("system", params, nil, false, tx)
	if err != nil {
		return nil, err
	}

	var bucketPolicies BucketPolicies

	for _, bp := range policies {
		b, err := BuildPolicy(bp)
		if err != nil {
			return nil, errors.New("critical, invalid policy: " + bp.Key["name"])
		}
		bucketPolicies = append(bucketPolicies, b)
	}

	sort.Slice(bucketPolicies, func(i, j int) bool { return bucketPolicies[i].Name < bucketPolicies[j].Name })

	return bucketPolicies, nil
}

// ApplyPolicies resolves Tags for a given bucket and key.
// 3 Tags are returned:
// - read : scope to add to objects you want to read
// - write : scope a user must have to write this object
// - promote : scope a user inherits for this bucket/key
func ApplyPolicies(policies BucketPolicies, bucket string, key Map, scope Tags) (bp BucketPolicies) {
	for _, p := range policies {
		if (p.Key.Contains(key) || key.Contains(p.Key)) && scope.Contains(p.Scope) {
			bp = append(bp, p)
		}
	}
	return bp
}
