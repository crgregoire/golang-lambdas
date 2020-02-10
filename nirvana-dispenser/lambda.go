package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

//
// LambdaMessage models the message
// sent from the lambda function
//
type LambdaMessage struct {
	Command     string    `json:"command"`
	Environment string    `json:"environment"`
	Timestamp   time.Time `json:"timestamp"`
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
	Flags              int    `json:"flags"`
	Inserted          *bool  `json:"inserted"`
	ServingsRemaining int    `json:"servingsRemaining"`
}

func handleRequest(ctx context.Context, lambdaMessage LambdaMessage) (string, error) {
	path := ""
	switch lambdaMessage.Command {
	case "dispensed":
		path = "/dispenser/dispensed"
	case "inserted":
		path = "/dispenser/inserted"
	case "connected":
		path = "/dispenser/connected"
	case "disconnected":
		path = "/dispenser/disconnected"
	}
	if strings.Contains(lambdaMessage.Payload.Dispenser.Serial, "dispenser-") {
		lambdaMessage.Payload.Dispenser.Serial = strings.Replace(lambdaMessage.Payload.Dispenser.Serial, "dispenser-", "", -1)
	}
	data, err := json.Marshal(lambdaMessage)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", os.Getenv("HOST")+path, bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("TOKEN"))

	client := &http.Client{}
	resp, err := client.Do(req)
	log.Print(resp.Body)
	if err != nil {
		panic(err)
	}
	return strconv.Itoa(resp.StatusCode), nil
}

func main() {
	lambda.Start(handleRequest)
}
