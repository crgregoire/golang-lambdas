package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/crgregoire/alexa"
	"github.com/tespo/satya/v2/types"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/service/iotdataplane"
)

func getUser(bearer string) (types.User, string) {
	user := types.User{}
	response, err := http.NewRequest("GET", os.Getenv("API_URL")+"/user", nil)
	response.Header.Add("Authorization", bearer)
	client := &http.Client{}
	resp, err := client.Do(response)
	if err != nil {
		return user, "I had trouble finding your profile"
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return user, "I had trouble finding your profile"
	}
	if err := json.Unmarshal(body, &user); err != nil {
		log.Println(err)
		return user, "I had trouble syncing your account"
	}
	return user, ""
}

func getDispenserSerial(bearer string) (string, string) {
	dispenserResponse, err := http.NewRequest("GET", os.Getenv("API_URL")+"/account/dispensers/", nil)
	dispenserResponse.Header.Add("Authorization", bearer)
	dispenserClient := &http.Client{}
	dispenserResp, err := dispenserClient.Do(dispenserResponse)
	if err != nil {
		return "", "It appears you don't have a dispenser linked to your Tespo Account"
	}
	dispenserBody, err := ioutil.ReadAll(dispenserResp.Body)
	if err != nil {
		return "", "It appears you don't have a dispenser linked to your Tespo Account"
	}

	var dispenser [10]byte
	var dispenserSerial string
	dispenserBytes := string([]byte(dispenserBody))
	x := 0
	for i := 64; i < 74; i++ {
		dispenser[x] = dispenserBytes[i]
		dispenserSerial += string(byte(dispenser[x]))
		x++
	}
	return dispenserSerial, ""
}

func shadowReturn(dispenserName string) (bool, string, int64, string){
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if err != nil {
		log.Print(err)
	}

	svc := iotdataplane.New(sess, &aws.Config{
		Endpoint: aws.String(os.Getenv("IOT_ENDPOINT")),
	})
	getShadowInput := iotdataplane.GetThingShadowInput{
		ThingName: &dispenserName,
	}
	thingShadow, err := svc.GetThingShadow(&getShadowInput)
	shadowBytes := string([]byte(thingShadow.Payload))
	inserted := gjson.Get(shadowBytes, "state.reported.payload.pod.inserted")
	servingsRemaining := gjson.Get(shadowBytes, "state.reported.payload.pod.servingsRemaining")
	barcode := gjson.Get(shadowBytes, "state.reported.payload.pod.barcode")
	barcodeString := barcode.String()
	servings := strconv.Itoa(int(servingsRemaining.Int()))
	servingsInt := servingsRemaining.Int()
	return inserted.Bool(), servings, servingsInt, barcodeString
}

func getPodString(barcodeString string) string {
	if barcodeString != "none" {
		barcodeResponse, err := http.NewRequest("GET", os.Getenv("POD_URL")+barcodeString+os.Getenv("MFG_TOKEN"), nil)
		barcodeClient := &http.Client{}
		barcodeResp, err := barcodeClient.Do(barcodeResponse)
		if err != nil {
			log.Print(err)
		}
		barcodeBody, err := ioutil.ReadAll(barcodeResp.Body)
		if err != nil {
			log.Print(err)
		}
		barcodeJSON := string([]byte(barcodeBody))
		podName := gjson.Get(barcodeJSON, "data.name")
		podString := podName.String()
		return podString
	} else {
		podString := ""
		return podString
	}
}

func handlePodSpeechReturn(insertion bool, servingsInt int64, podString string, servings string) alexa.Response {
	var builder alexa.SSMLBuilder
	if insertion {
		if servingsInt > 1 {
			builder.Say("You currently have a " + podString + " pod inserted and it has " + servings + " servings left!")
		} else if servingsInt == 1 {
			builder.Say("You currently have a " + podString + " pod inserted and it has " + servings + " serving left!")
		} else {
			builder.Say("You currently have a " + podString + " pod inserted,")
			builder.Say("But, your " + podString + " pod is out of servings! Please insert a new pod, or order one at get<break time='75ms'/>tespo.com!")
		}
	} else {
		builder.Say("You do not currently have a pod inserted.")
		builder.Pause("250")
		builder.Say("Please insert a pod to use this functionality!")
	}
	builder.Pause("500")
	return alexa.NewSSMLResponse("Your Tespo Information", builder.Build())
}

func deviceSpeechReturn(user types.User, dispenserSerial string, insertion bool, servingsInt int64, podString string, servings string) alexa.Response {
	rand.Seed(time.Now().UnixNano())
	randomPhrase := rand.Intn(6)
	var builder alexa.SSMLBuilder
	builder.Say("Here is the info regarding your dispenser.")
	builder.Pause("500")
	builder.Say("Here's your dispenser name: dispenser-<say-as interpret-as=\"spell-out\">" + dispenserSerial + "</say-as>, and serial number: <say-as interpret-as=\"spell-out\">" + dispenserSerial + "</say-as>")
	builder.Pause("500")
	builder.Say("This dispenser is linked to the account belonging to " + user.FirstName + " " + user.LastName + ".")
	builder.Pause("500")
	if insertion {
		if servingsInt > 0 {
			builder.Say("Your " + podString + " pod is currently inserted and has " + servings + " servings left!")

		} else {
			builder.Say("Unfortunately, your " + podString + " pod is out of servings! Please insert a new pod, or order one at get<break time='75ms'/>tespo.com!")
		}
	} else {
		builder.Say("Please insert a pod in order to dispense!")
	}
	builder.Pause("500")
	switch randomPhrase {
	case 0:
		builder.Say("Tespo,<break time='75ms'/> A Daily Dose Of Shine!")
		break
	case 1:
		builder.Say("Tespo,<break time='75ms'/> Discover A New Wellness!")
		break
	case 2:
		builder.Say("Did you know that Tespo offers fs on all products?")
		builder.Pause("250")
		builder.Say("Something to think about. Have a healthy day!")
		break
	case 3:
		builder.Say("Tespo,<break time='75ms'/> Something To Feel Good About!")
		break
	case 4:
		builder.Say("Tespo.<break time='75ms'/> Your Health,<break time='75ms'/> Our Mission!")
		break
	default:
		builder.Say("Tespo,<break time='75ms'/> Wellness That Works!")
	}

	builder.Pause("500")
	return alexa.NewSSMLResponse("Your Tespo Information", builder.Build())
}

//
// HandlePodIntent returns a response about the user's currently inserted pod
//
func HandlePodIntent(request alexa.Request) alexa.Response {
	token := request.Session.User.AccessToken
	if token == "" {
		return alexa.NewLinkAccountResponse()
	}
	var bearer = "Bearer " + token

	dispenserSerial, errString := getDispenserSerial(bearer)
	if errString != "" {
		return alexa.NewSimpleResponse("Dispenser Not Found", "It appears you don't have a dispenser linked to your Tespo Account")
	}

	dispenserName := "dispenser-" + dispenserSerial
	insertion, servings, servingsInt, barcodeString := shadowReturn(dispenserName)

	podString := getPodString(barcodeString)

	return handlePodSpeechReturn(insertion, servingsInt, podString, servings)
}

//
// HandleDeviceIntent returns a speech representation of the user's dispenser and inserted pod if any
//
func HandleDeviceIntent(request alexa.Request) alexa.Response {
	token := request.Session.User.AccessToken
	if token == "" {
		return alexa.NewLinkAccountResponse()
	}
	var bearer = "Bearer " + token

	user, errorString := getUser(bearer)
	if errorString != "" {
		return alexa.NewSimpleResponse("Account Error", errorString)
	}

	dispenserSerial, errString := getDispenserSerial(bearer)
	if errString != "" {
		return alexa.NewSimpleResponse("Dispenser Not Found", errString)
	}

	dispenserName := "dispenser-" + dispenserSerial
	insertion, servings, servingsInt, barcodeString := shadowReturn(dispenserName)

	podString := getPodString(barcodeString)

	return deviceSpeechReturn(user, dispenserSerial, insertion, servingsInt, podString, servings)
}

func HandleDispenseIntent(request alexa.Request) alexa.Response {
	token := request.Session.User.AccessToken
	if token == "" {
		return alexa.NewLinkAccountResponse()
	}
	var bearer = "Bearer " + token
	response, err := http.NewRequest("GET", os.Getenv("API_URL")+"/user", nil)
	response.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(response)
	if err != nil {
		return alexa.NewSimpleResponse("Account Error", "I had trouble finding your profile")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return alexa.NewSimpleResponse("Account Error", "I had trouble finding your profile")
	}
	user := types.User{}
	if err := json.Unmarshal(body, &user); err != nil {
		log.Println(err)
	}

	dispenserResponse, err := http.NewRequest("GET", os.Getenv("API_URL")+"/account/dispensers/", nil)
	dispenserResponse.Header.Add("Authorization", bearer)

	dispenserClient := &http.Client{}
	dispenserResp, err := dispenserClient.Do(dispenserResponse)
	if err != nil {
		return alexa.NewSimpleResponse("Dispenser Not Found", "It appears you don't have a dispenser linked to your Tespo Account")
	}
	dispenserBody, err := ioutil.ReadAll(dispenserResp.Body)
	if err != nil {
		return alexa.NewSimpleResponse("Dispenser Not Found", "It appears you don't have a dispenser linked to your Tespo Account")
	}

	var dispenser [10]byte
	var dispenserSerial string
	dispenserBytes := string([]byte(dispenserBody))
	x := 0
	for i := 64; i < 74; i++ {
		dispenser[x] = dispenserBytes[i]
		dispenserSerial += string(byte(dispenser[x]))
		x++
	}

	dispenserName := "dispenser-" + dispenserSerial
	payload := types.Payload{}
	payload.Customer.ID = user.AccountID.String()
	payload.Dispenser.Serial = dispenserSerial
	payload.Dispenser.Name = dispenserName

	lambdaMessage := types.LambdaMessage{}
	lambdaMessage.Payload.Dispenser.Name = dispenserName
	lambdaMessage.Payload.Dispenser.Serial = dispenserSerial
	lambdaMessage.Payload.Customer.ID = user.ID.String()
	lambdaMessage.Timestamp = time.Now()
	lambdaMessage.Command = "dispense"

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if err != nil {
		log.Print(err)
	}
	log.Print(lambdaMessage)
	svc1 := iotdataplane.New(sess, &aws.Config{
		Endpoint: aws.String(os.Getenv("IOT_ENDPOINT")),
	})
	getShadowInput := iotdataplane.GetThingShadowInput{
		ThingName: &dispenserName,
	}
	thingShadow, err := svc1.GetThingShadow(&getShadowInput)
	shadowBytes := string([]byte(thingShadow.Payload))
	inserted := gjson.Get(shadowBytes, "state.reported.payload.pod.inserted")
	servingsRemaining := gjson.Get(shadowBytes, "state.reported.payload.pod.servingsRemaining")
	barcode := gjson.Get(shadowBytes, "state.reported.payload.pod.barcode")
	barcodeString := barcode.String()
	var podString string
	if barcodeString != "none" {
		barcodeResponse, err := http.NewRequest("GET", os.Getenv("POD_URL")+barcodeString+os.Getenv("MFG_TOKEN"), nil)

		barcodeClient := &http.Client{}
		barcodeResp, err := barcodeClient.Do(barcodeResponse)
		if err != nil {
			log.Print(err)
		}
		barcodeBody, err := ioutil.ReadAll(barcodeResp.Body)
		if err != nil {
			log.Print(err)
		}
		barcodeJSON := string([]byte(barcodeBody))
		podName := gjson.Get(barcodeJSON, "data.name")
		podString = podName.String()
	} else {
		podString = ""
	}
	var servings string
	var servingsInt int64
	var insertion bool
	servings = strconv.Itoa(int(servingsRemaining.Int()))
	servingsInt = servingsRemaining.Int()
	log.Print(servings)
	insertion = inserted.Bool()
	var builder alexa.SSMLBuilder
	if insertion {
		if servingsInt > 0 {
			builder.Say("I'm dispensing your " + podString + " pod now!")
			builder.Pause("500")
			data, err := json.Marshal(lambdaMessage)
			publishInput := iotdataplane.PublishInput{
				Payload: data,
				Topic:   aws.String("incoming-cmd/" + lambdaMessage.Payload.Dispenser.Name),
			}
			svc := iotdataplane.New(sess, &aws.Config{
				Endpoint: aws.String(os.Getenv("IOT_ENDPOINT")),
			})
			out, err := svc.Publish(&publishInput)
			out2, err := svc.Publish(&publishInput)
			log.Print(out)
			log.Print(out2)
			if err != nil {
				log.Print(err)
				return alexa.NewSimpleResponse("Error with Dispensing", "I had trouble dispensing your serving! Common problems include disconnected Wi-Fi, lower water level, and no pod inserted.'")
			}
			return alexa.NewSSMLResponse("Serving Dispensed!", builder.Build())

		} else {
			builder.Say("You currently have a " + podString + " pod inserted,")
			builder.Say("But, your " + podString + " pod is out of servings! Please insert a new pod, or order one at get<break time='75ms'/>tespo.com!")
		}
	} else {
		builder.Say("You do not currently have a pod inserted.")
		builder.Pause("250")
		builder.Say("Please insert a pod to dispense your serving!")
	}
	return alexa.NewSSMLResponse("Error Dispensing", builder.Build())
}

func HandleHelpIntent(request alexa.Request) alexa.Response {
	var builder alexa.SSMLBuilder
	builder.Say("Here are some of the things you can ask:")
	builder.Pause("1000")
	builder.Say("Alexa, ask my dispenser to dispense a serving.")
	builder.Pause("1000")
	builder.Say("or, Alexa, ask my Dispenser about this skill.")
	builder.Pause("1000")
	builder.Say("or, Alexa, ask my Dispenser what pod is inserted?")
	builder.Pause("1000")
	builder.Say("or, Alexa, ask my Dispenser to tell me about itself.")
	builder.Pause("1000")
	builder.Say("or, Alexa, ask my Dispenser for help.")
	return alexa.NewSSMLResponse("Tespo Connect Help", builder.Build())
}

func HandleAboutIntent(request alexa.Request) alexa.Response {
	rand.Seed(time.Now().UnixNano())
	randomPhrase := rand.Intn(6)
	var builder alexa.SSMLBuilder
	builder.Say("Tespo is a better way to take a better vitamin!")
	builder.Pause("500")
	builder.Say("We designed the first vitamin dispenser that mixes raw, powdered ingredients into a fresh, liquid shot.")
	builder.Pause("500")
	switch randomPhrase {
	case 0:
		builder.Say("Start your day with something good. Start your day with Tespo. ")
		break
	case 1:
		builder.Say("Tespo, Helping You Shine!")
		break
	case 2:
		builder.Say("Tespo, Discover A New Wellness!")
		break
	case 3:
		builder.Say("Tespo, For The Health Of It!")
		break
	case 4:
		builder.Say("Tespo. Keeping You Well!")
		break
	default:
		builder.Say("Tespo, Clean & Convenient Care!")
	}
	builder.Say("To dispense a serving, say, \"Alexa, ask my dispenser to dispense a serving\"")
	return alexa.NewSSMLResponse("About", builder.Build())
}

func IntentDispatcher(request alexa.Request) alexa.Response {
	var response alexa.Response
	switch request.Body.Intent.Name {
	case "PodIntent":
		response = HandlePodIntent(request)
		break
	case "DispenseIntent":
		response = HandleDispenseIntent(request)
		break
	case "DeviceIntent":
		response = HandleDeviceIntent(request)
		break
	case alexa.HelpIntent:
		response = HandleHelpIntent(request)
		break
	case "AboutIntent":
		response = HandleAboutIntent(request)
		break
	default:
		response = HandleHelpIntent(request)
	}
	return response
}

func Handler(request alexa.Request) (alexa.Response, error) {
	return IntentDispatcher(request), nil
}

func main() {
	lambda.Start(Handler)
}
