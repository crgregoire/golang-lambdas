package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

// Dispenser event example
// {
//     "command":"dispensed",
//     "environment":"production",
//     "payload": {
//         "customer":{
//             "id":1,
//         },
//         "dispenser":{
//             "name":"test",
//             "serial":"test-serial",
//             "controllerFirmwareVersion":"version",
//             "wifiFirmwareVersion":"version",
//             "pcbFirmwareVersion":"version",
//         },
//         "pod":{
//             "Barcode":"00400000000002",
//             "Flag":"test",
//             "Inserted":"test",
//             "ServingsRemaining":"test",
//         }
//     }
// }

type request struct {
	Name              string `json:"name"`
	SKU               string `json:"sku"`
	CustomerID        string `json:"customer_id"`
	Barcode           string `json:"barcode"`
	ServingsRemaining int    `json:"servings_remaining"`
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
	Flag              string `json:"flag"`
	Inserted          string `json:"inserted"`
	ServingsRemaining int    `json:"servingsRemaining"`
}

type event struct {
	Command     string    `json:"command"`
	Environment string    `json:"environment"`
	Timestamp   time.Time `json:"timestamp"`
	Payload     payload   `json:"payload"`
}

type apiPod struct {
	Data struct {
		Name            string   `json:"name"`
		SKU             string   `json:"sku"`
		Color           string   `json:"-"`
		NutritionLabels []string `json:"-"`
	} `json:"data"`
}

var ids = []string{
	"0",
	"1",
	"2",
	"3",
	"4",
	"5",
	"6",
	"7",
	"8",
	"9",
	"10",
	"11",
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func getPodDetails(barcode string) (apiPod, error) {
	pod := apiPod{}
	req, err := http.NewRequest("GET", os.Getenv("MFG_API_ENDPOINT")+"/api/pods/"+barcode, nil)
	if err != nil {
		return pod, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("MFG_API_TOKEN"))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print(err)
		return pod, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		return pod, err
	}
	if err := json.Unmarshal(data, &pod); err != nil {
		return pod, err
	}
	if pod.Data.SKU != "LB8000-006" {
		return pod, errors.New("Pod is not a sleep pod: " + pod.Data.SKU)
	}

	return pod, nil
}

func handleRequest(ctx context.Context, e event) (string, error) {
	pod, err := getPodDetails(e.Payload.Pod.Barcode)

	if err != nil {
		if strings.Contains(err.Error(), "Pod is not a sleep pod") {
			return err.Error(), nil
		}
		return err.Error(), err
	}

	var request = request{
		Name:              pod.Data.Name,
		SKU:               pod.Data.SKU,
		CustomerID:        e.Payload.Customer.ID,
		Barcode:           e.Payload.Pod.Barcode,
		ServingsRemaining: e.Payload.Pod.ServingsRemaining,
	}

	data, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", os.Getenv("BIOMARKER_ENDPOINT"), bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		panic(err)
	}
	return "success", nil
}

func main() {
	lambda.Start(handleRequest)
}
