package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iotdataplane"
	"github.com/tespo/satya/v2/types"
)

func handleRequest(ctx context.Context, e map[string]interface{}) (*iotdataplane.PublishOutput, error) {

	d, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	lambdaMessage := types.LambdaMessage{}
	if err := json.Unmarshal(d, &lambdaMessage); err != nil {
		return nil, err
	}
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	publishInput := iotdataplane.PublishInput{
		Payload: data,
		Topic:   aws.String("incoming-cmd/" + lambdaMessage.Payload.Dispenser.Name),
	}
	svc := iotdataplane.New(sess, &aws.Config{
		Endpoint: aws.String(os.Getenv("IOT_ENDPOINT")),
	})
	out, err := svc.Publish(&publishInput)
	if err != nil {
		return nil, err
	}
	out, err = svc.Publish(&publishInput)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func main() {
	lambda.Start(handleRequest)
}
