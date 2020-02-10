package main

import (
	"context"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"
)

//
// LambdaMessage models the message
// sent from the lambda function
//
type LambdaMessage struct {
	Command     string    `json:"command"`
	Environment string    `json:"environment"`
	Timestamp   string `json:"timestamp"`
	Payload     payload   `json:"payload"`
}

type customer struct {
	ID string `json:"id"`
}

type payload struct {
	Customer  customer  `json:"customer"`
	Dispenser dispenser `json:"dispenser"`
	Pod       pod       `json:"pod"`
}

type dispenser struct {
	Name                      string `json:"name"`
	Serial                    string `json:"serial"`
	ControllerFirmwareVersion string `json:"controllerFirmwareVersion"`
	WifiFirmwareVersion       string `json:"wifiFirmwareVersion"`
	PcbFirmwareVersion        string `json:"pcbFirmwareVersion"`
}

type pod struct {
	Barcode           string `json:"barcode"`
	Flag              int    `json:"flag"`
	Inserted          *bool  `json:"inserted"`
	ServingsRemaining int    `json:"servingsRemaining"`
}

func handleRequest(ctx context.Context, lambdaMessage LambdaMessage) (string, error) {
	path := "/wp-json/wc-tespo/v1/reward-user-usage?consumer_key=" + os.Getenv("CONSUMER_KEY") + "&consumer_secret=" + os.Getenv("CONSUMER_SECRET") + "&account_id=" + lambdaMessage.Payload.Customer.ID

	req, err := http.NewRequest("POST", os.Getenv("HOST")+path, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	return strconv.Itoa(resp.StatusCode), nil
}

func main() {
	lambda.Start(handleRequest)
}
