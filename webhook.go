package webhook 

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type intent struct {
	DisplayName string `json:"displayName"`
}

type queryResult struct {
	Intent intent `json:"intent"`
}

type text struct {
	Text []string `json:"text"`
}

type message struct {
	Text text `json:"text"`
}

// webhookRequest is used to unmarshal a WebhookRequest JSON object. Note that
// not all members need to be defined--just those that you need to process.
// As an alternative, you could use the types provided by
// the Dialogflow protocol buffers:
// https://godoc.org/google.golang.org/genproto/googleapis/cloud/dialogflow/v2#WebhookRequest
type webhookRequest struct {
	Session     string      `json:"session"`
	ResponseID  string      `json:"responseId"`
	QueryResult queryResult `json:"queryResult"`
}

// webhookResponse is used to marshal a WebhookResponse JSON object. Note that
// not all members need to be defined--just those that you need to process.
// As an alternative, you could use the types provided by
// the Dialogflow protocol buffers:
// https://godoc.org/google.golang.org/genproto/googleapis/cloud/dialogflow/v2#WebhookResponse
type webhookResponse struct {
	FulfillmentMessages []message `json:"fulfillmentMessages"`
}

// welcome creates a response for the welcome intent.
func welcome(request webhookRequest) (webhookResponse, error) {
	response := webhookResponse{
		FulfillmentMessages: []message{
			{
				Text: text{
					Text: []string{"Welcome from Dialogflow Go Webhook"},
				},
			},
		},
	}
	return response, nil
}

// getAgentName creates a response for the get-agent-name intent.
func getAgentName(request webhookRequest) (webhookResponse, error) {
	response := webhookResponse{
		FulfillmentMessages: []message{
			{
				Text: text{
					Text: []string{"My name is Dialogflow Go Webhook"},
				},
			},
		},
	}
	return response, nil
}

// handleError handles internal errors.
func handleError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "ERROR: %v", err)
}

// HandleWebhookRequest handles WebhookRequest and sends the WebhookResponse.
func HandleWebhookRequest(w http.ResponseWriter, r *http.Request) {
	var request webhookRequest
	var response webhookResponse
	var err error

	// Read input JSON
	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		handleError(w, err)
		return
	}
	log.Printf("Request: %+v", request)

	// Call intent handler
	switch intent := request.QueryResult.Intent.DisplayName; intent {
	case "Default Welcome Intent":
		response, err = welcome(request)
	case "get-agent-name":
		response, err = getAgentName(request)
	default:
		err = fmt.Errorf("Unknown intent: %s", intent)
	}
	if err != nil {
		handleError(w, err)
		return
	}
	log.Printf("Response: %+v", response)

	// Send response
	if err = json.NewEncoder(w).Encode(&response); err != nil {
		handleError(w, err)
		return
	}
}
