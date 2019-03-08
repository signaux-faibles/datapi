package dalib

import (
	"fmt"
	"regexp"
	"sort"
	"sync"
	"time"
)

var policiesLock sync.Mutex

// BucketPolicies is a slice of policies with apply method
type BucketPolicies []BucketPolicy

func (p BucketPolicies) safeRead() BucketPolicies {
	policiesLock.Lock()
	defer policiesLock.Unlock()
	return p
}

func (p *BucketPolicies) safeUpdate(np BucketPolicies) {
	policiesLock.Lock()
}

// CurrentBucketPolicies is a cache object
var CurrentBucketPolicies BucketPolicies

// LoadPolicies returns bucket policies at a given time
func LoadPolicies(date *time.Time) (BucketPolicies, error) {
	var params QueryParams
	params.Key = Map{
		"type": "policy",
	}
	params.Date = date

	policies, err := Query("auth", params, Tags{"auth", "admin"})
	if err != nil {
		return nil, err
	}

	var bucketPolicies BucketPolicies

	for _, bp := range policies {
		match := bp.Key["match"]
		read, err := ToScope(bp.Value["read"].([]interface{}))
		if err != nil {
			return nil, err
		}
		write, err := ToScope(bp.Value["write"].([]interface{}))
		if err != nil {
			return nil, err
		}
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

	return bucketPolicies, nil
}
