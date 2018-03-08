package main

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/wolfeidau/lambda-cache-example/pkg/ssmcache"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const (
	defaultSSMKey = "/mwolfe/caching-example"
)

type server struct {
	cache ssmcache.Cache
}

func newServer() *server {
	sess := session.New()

	ssmCache := ssmcache.New(sess)

	return &server{
		cache: ssmCache,
	}
}

func (cs *server) handler(request events.SNSEvent) error {

	val, err := cs.cache.GetKey(defaultSSMKey)
	if err != nil {
		log.Fatalf("Error retrieving cached resource: %+v", err)
	}

	log.Println("key:", defaultSSMKey, "value:", val)

	return nil
}

func main() {
	log.Println("cold start at:", time.Now())
	cs := newServer()
	lambda.Start(cs.handler)
}
