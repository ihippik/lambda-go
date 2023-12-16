# Lambda-Go

[![Publish Docker image](https://github.com/ihippik/lambda-go/actions/workflows/github-actions.yml/badge.svg)](https://github.com/ihippik/lambda-go/actions/workflows/github-actions.yml)

This project is an attempt to understand how the most popular serverless computing solution, Lambda functions, works.

It was inspired by the most popular solution, [AWS Lambda](https://aws.amazon.com/lambda/).

## How it works

At this stage, the project is a web server with two endpoints.
Server port should be specified in the `APP_SERVER_PORT` environment variable.

Also we need to specify logs level in the `APP_LOG_LEVEL` environment variable.

For proper operation, the server must have access to the ***Docker*** daemon, 
which is used to deploy our functions in containers.

### Create function

This endpoint allows us to upload an archive with Lambda function code to the server.

```shell
curl --location 'localhost:9000/lambda/{func_name}/create' \
--form 'file=@"/func.tar.gz"'
```

To create a new Lambda function, you need to specify the function name `{func_name}` and a tar.gz archive.

The archive must contain three files:
* main.go
* go.mod
* go.sum

The code for the application should look like the following:
    
```go
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/ihippik/lambda-go/lambda"
)

type Data struct {
	Name string `json:"name"`
}

func hello(_ context.Context, data []byte) ([]byte, error) {
	var req Data

	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&req); err != nil {
		return nil, err
	}

	return []byte(fmt.Sprintf("Hello %s!", req.Name)), nil
}

func main() {
	lambda.Start(hello)
}
```

All handlers must have the following signature:

```go
type Handler func(ctx context.Context, payload []byte) ([]byte, error)
```

To prepare the archive, you can use the following command:

```shell 
tar -czvf func.tar.gz main.go go.mod go.sum
```

### Invoke function

This endpoint allows us to run a previously uploaded function `{func_name}`

```shell
curl --location 'localhost:9000/lambda/{func_name}/invoke' \
--header 'Content-Type: application/json' \
--data '{"name": "Ivan"}'
```
