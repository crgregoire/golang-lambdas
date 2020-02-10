package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tespo/satya/v2/types"
)

func handleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	response := events.APIGatewayProxyResponse{}
	updateUser := types.User{}
	if err := json.Unmarshal([]byte(request.Body), &updateUser); err != nil {
		response.Body = err.Error()
		return response, err
	}
	sendData, err := json.Marshal(updateUser)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("PUT", os.Getenv("UPDATE_USER_BY_EXTERNAL_ID_ENDPOINT")+"/user/"+strconv.Itoa(int(updateUser.ExternalID)), bytes.NewReader(sendData))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("TOKEN"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode > 200 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		response.Body = string(data)
		response.StatusCode = resp.StatusCode
		return response, nil
	}

	response.Body = "success"
	response.StatusCode = 200
	return response, nil
}

func main() {
	lambda.Start(handleRequest)
}
