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
	Name    string
	Match   func(s string) bool // apply the policy if Match(bucket's name) == true
	Scope   Tags                // user needs this scope to apply policy
	Key     Map                 // apply this policy for objects with this subkey
	Read    Tags                // user needs this scope to read objects
	Write   Tags                // user needs this scope to write objects
	Promote Tags                // user inherits this scope from policy
}

// BucketPolicies is a slice of policies with apply method
type BucketPolicies []BucketPolicy

func (p BucketPolicies) safeRead() BucketPolicies {
	policiesLock.Lock()
	defer policiesLock.Unlock()
	return p
}

func (p *BucketPolicies) safeUpdate(np BucketPolicies) {
	policiesLock.Lock()
	defer policiesLock.Unlock()
	*p = np
}

// CurrentBucketPolicies is a cache object
var CurrentBucketPolicies BucketPolicies

// BuildPolicy translates Object to BucketPolicy
func BuildPolicy(o Object) (BucketPolicy, error) {
	name, ok := o.Key["name"]
	if !ok {
		return BucketPolicy{}, errors.New("name subkey is mandatory for policies")
	}
	match := o.Value["match"].(string)
	read, err := ToScope(o.Value["read"])
	if err != nil {
		return BucketPolicy{}, err
	}
	write, err := ToScope(o.Value["write"])
	if err != nil {
		return BucketPolicy{}, err
	}
	key, err := ToMap(o.Value["key"])
	if err != nil {
		return BucketPolicy{}, err
	}

	matchre := regexp.MustCompile(match)

	bp := BucketPolicy{
		Name:  name,
		Match: matchre.MatchString,
		Read:  read,
		Key:   key,
		Write: write,
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
func ApplyPolicies(policies BucketPolicies, bucket string, key Map, scope Tags) (read Tags, write Tags, promote Tags) {
	readtags := make(map[string]struct{})
	promotetags := make(map[string]struct{})
	writetags := make(map[string]struct{})

	for _, p := range policies {
		if scope.Contains(p.Scope) {
			if p.Match(bucket) && key.contains(p.Key) {
				for _, tag := range p.Read {
					readtags[tag] = struct{}{}
				}
				for _, tag := range p.Write {
					writetags[tag] = struct{}{}
				}
				for _, tag := range p.Promote {
					promotetags[tag] = struct{}{}
				}
			}
		}
	}

	for tag := range readtags {
		read = append(read, tag)
	}
	for tag := range writetags {
		write = append(read, tag)
	}
	for tag := range promotetags {
		promote = append(read, tag)
	}

	return read, write, promote
}
