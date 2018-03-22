package ssmcache

import (
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/pkg/errors"
)

var defaultExpiry = 30 * time.Second

// SetDefaultExpiry update the default expiry for all cached parameters
//
// Note this will update expires value on the next refresh of entries.
func SetDefaultExpiry(expires time.Duration) {
	defaultExpiry = expires
}

// Entry an SSM entry in the cache
type Entry struct {
	value   string
	expires time.Time
}

// Cache SSM cache which provides read access to parameters
type Cache interface {
	GetKey(key string) (string, error)
}

type cache struct {
	ssm       sync.Mutex
	ssmValues map[string]*Entry
	ssmSvc    ssmiface.SSMAPI
}

// New new SSM cache
func New(sess *session.Session) Cache {
	return &cache{
		ssmSvc:    ssm.New(sess),
		ssmValues: make(map[string]*Entry),
	}
}

// GetKey retrieve a parameter from SSM and cache it.
func (ssc *cache) GetKey(key string) (string, error) {

	ssc.ssm.Lock()
	defer ssc.ssm.Unlock()

	ent, ok := ssc.ssmValues[key]
	if !ok {
		// record is missing
		return ssc.updateParam(key)
	}

	if time.Now().After(ent.expires) {
		// we have expired and need to refresh
		log.Println("expired cache refreshing value")

		return ssc.updateParam(key)
	}

	// return the value
	return ent.value, nil
}

func (ssc *cache) updateParam(key string) (string, error) {

	log.Println("updating key from ssm:", key)

	resp, err := ssc.ssmSvc.GetParameter(&ssm.GetParameterInput{
		Name: aws.String(key),
	})
	if err != nil {
		return "", errors.Wrapf(err, "failed to retrieve key %s from ssm", key)
	}

	ssc.ssmValues[key] = &Entry{
		value:   aws.StringValue(resp.Parameter.Value),
		expires: time.Now().Add(defaultExpiry), // reset the expiry
	}

	log.Println("key value refreshed from ssm at:", time.Now())

	return aws.StringValue(resp.Parameter.Value), nil
}
