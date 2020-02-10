package main

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codepipeline"
)

type contentfulWebhookBody struct {
	Sys sys `json:"sys"`
}

type sys struct {
	Environment environment `json:"environment"`
}

type environment struct {
	Sys environmentSys `json:"sys"`
}
type environmentSys struct {
	ID string `json:"id"`
}

func handleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	response := events.APIGatewayProxyResponse{}
	contentfulWebhookBody := contentfulWebhookBody{}
	if err := json.Unmarshal([]byte(request.Body), &contentfulWebhookBody); err != nil {
		response.Body = err.Error()
		return response, err
	}
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if err != nil {
		response.Body = err.Error()
		return response, err
	}
	svc := codepipeline.New(sess)
	_, err = svc.StartPipelineExecution(&codepipeline.StartPipelineExecutionInput{
		ClientRequestToken: aws.String("contentful-update-" + strconv.Itoa(int(time.Now().Unix()))),
		Name:               aws.String("vitta-" + contentfulWebhookBody.Sys.Environment.Sys.ID),
	})
	if err != nil {
		response.Body = err.Error()
		return response, err
	}
	response.Body = "success"
	response.StatusCode = 200
	return response, nil
}

func main() {
	lambda.Start(handleRequest)
}
