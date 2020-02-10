package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tespo/satya/v2/types"
)

func handleRequest(ctx context.Context, e map[string]interface{}) (*types.AlexaResponse, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", os.Getenv("ALEXA_FULFILLMENT_ENDPOINT"), bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	decoder := json.NewDecoder(resp.Body)
	buddhaResponse := types.AlexaResponse{}
	if err := decoder.Decode(&buddhaResponse); err != nil {
		return nil, errors.New("Something didn't go right")
	}
	return &buddhaResponse, nil
}

func main() {
	lambda.Start(handleRequest)
}
