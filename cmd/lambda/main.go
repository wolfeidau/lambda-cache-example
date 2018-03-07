package main

import (
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/pkg/errors"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const (
	defaultExpiry = 30 * time.Second
	defaultSSMKey = "/mwolfe/caching-example"
)

type cachingServer struct {
	ssm      sync.Mutex
	ssmValue string
	expires  time.Time
	ssmSvc   ssmiface.SSMAPI
}

func newCachingServer() *cachingServer {
	sess := session.New()

	return &cachingServer{
		ssmSvc:  ssm.New(sess),
		expires: time.Now(), // set the initial expiry to now
	}
}

func (cs *cachingServer) handler(request events.SNSEvent) error {

	err := cs.checkExpires(defaultSSMKey)
	if err != nil {
		log.Fatalf("Error retrieving cached resource: %+v", err)
	}

	return nil
}

func (cs *cachingServer) checkExpires(key string) error {

	cs.ssm.Lock()
	defer cs.ssm.Unlock()

	if time.Now().After(cs.expires) {
		// we have expired and need to refresh
		log.Println("expired cache refreshing value")

		err := cs.updateParam(key)
		if err != nil {
			return errors.Wrap(err, "failed to update param")
		}
	}

	return nil
}

func (cs *cachingServer) updateParam(key string) error {

	log.Println("updating key from ssm:", defaultSSMKey)

	resp, err := cs.ssmSvc.GetParameter(&ssm.GetParameterInput{
		Name: aws.String(defaultSSMKey),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve key %s from ssm", defaultSSMKey)
	}

	cs.ssmValue = aws.StringValue(resp.Parameter.Value)

	log.Println("key value refreshed from ssm at:", time.Now())

	cs.expires = time.Now().Add(defaultExpiry) // reset the expiry

	return nil
}

func main() {
	log.Println("cold start at:", time.Now())
	cs := newCachingServer()
	lambda.Start(cs.handler)
}
