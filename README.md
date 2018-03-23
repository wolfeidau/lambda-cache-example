# lambda-cache-example

This project provides an example [golang](https://golang.org) based lambda function which caches a key from parameter store in a local variable and refreshes it every 30 seconds. This ensures that you only hit the [AWS Systems Manager Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-paramstore.html) API to refresh and not every time the lambda is triggered, therefore avoiding rate limiting your self.

# example 

The below example is based on the api gateway code in [Announcing Go Support for AWS Lambda](https://aws.amazon.com/blogs/compute/announcing-go-support-for-aws-lambda/).

```go
package main

import (
	"errors"
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/wolfeidau/lambda-cache-example/pkg/ssmcache"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const defaultSSMKey = "/mwolfe/caching-example"

var (
	errNameNotProvided  = errors.New("no name was provided in the HTTP body")
	errConfigLoadFailed = errors.New("unable to load configuration")

	cache = ssmcache.New(session.New())
)

// Handler is your Lambda function handler
// It uses Amazon API Gateway request/responses provided by the aws-lambda-go/events package,
// However you could use other event sources (S3, Kinesis etc), or JSON-decoded primitive types such as 'string'.
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// stdout and stderr are sent to AWS CloudWatch Logs
	log.Printf("Processing Lambda request %s\n", request.RequestContext.RequestID)

	// If no name is provided in the HTTP request body, return an error
	if len(request.Body) < 1 {
		return events.APIGatewayProxyResponse{}, errNameNotProvided
	}

	val, err := cache.GetKey(defaultSSMKey)
	// if we failed to load configuration return an error
	if err != nil {
		return events.APIGatewayProxyResponse{}, errConfigLoadFailed
	}

	log.Printf("key: %s value: %s", defaultSSMKey, val)

	return events.APIGatewayProxyResponse{
		Body:       "Hello " + request.Body,
		StatusCode: 200,
	}, nil

}

func main() {
	lambda.Start(Handler)
}
```

# license

This code is released under MIT License, and is copyright Mark Wolfe.