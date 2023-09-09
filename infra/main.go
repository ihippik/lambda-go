package main

import (
	"context"

	"github.com/ihippik/lambda-go/lambda"
)

func hello(ctx context.Context, data []byte) ([]byte, error) {
	return data, nil
}

func main() {
	// Make the handler available for Remote Procedure Call by AWS Lambda
	lambda.Start(hello)
}
