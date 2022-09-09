package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"cloud.google.com/go/translate"
	"golang.org/x/text/language"
)

type intent struct {
	DisplayName string `json:"displayName"`
}

type location struct {
	City string `json:"city"`
}

type parameters struct {
	Location location `json:"location"`
}

type queryResult struct {
	Intent     intent     `json:"intent"`
	Parameters parameters `json:"parameters"`
}

type text struct {
	Text []string `json:"text"`
}

type message struct {
	Text text `json:"text"`
}

type locationMessages struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type weather struct {
	Main        string `json:"main"`
	Description string `json:"description"`
}

type weatherResponse struct {
	Weather []weather `json:"weather"`
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

// detectLanguage translate the language of city
func detectLanguage(text string) (*translate.Detection, error) {
	// text := "こんにちは世界"
	ctx := context.Background()
	client, err := translate.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("translate.NewClient: %v", err)
	}
	defer client.Close()
	lang, err := client.DetectLanguage(ctx, []string{text})
	if err != nil {
		return nil, fmt.Errorf("DetectLanguage: %v", err)
	}
	if len(lang) == 0 || len(lang[0]) == 0 {
		return nil, fmt.Errorf("DetectLanguage return value empty")
	}
	return &lang[0][0], nil
}

// translate text string
func translateText(targetLanguage, text string) (string, error) {
	// text := "The Go Gopher is cute"
	ctx := context.Background()

	lang, err := language.Parse(targetLanguage)
	if err != nil {
		return "", fmt.Errorf("language.Parse: %v", err)
	}

	client, err := translate.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	resp, err := client.Translate(ctx, []string{text}, lang, nil)
	if err != nil {
		return "", fmt.Errorf("Translate: %v", err)
	}
	if len(resp) == 0 {
		return "", fmt.Errorf("Translate returned empty response to text: %s", text)
	}
	return resp[0].Text, nil
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

// weather creates a response for the weather intent.
func weatherQuery(request webhookRequest) (webhookResponse, error) {
	api_key := "xxxxxxxxxxx"
	city := request.QueryResult.Parameters.Location.City
	url := "http://api.openweathermap.org/geo/1.0/direct?q=" + city + "&limit=1&appid=" + api_key
	var locationmessages []locationMessages
	var lat float64
	var lon float64
	var weatherresponse weatherResponse
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(body, &locationmessages); err != nil {
		log.Fatal(err)
	}

	//	for l := range locationmessages {
	//		fmt.Printf("Lat = %v, Lon = %v", locationmessages[l].Lat, locationmessages[l].Lon)
	//		fmt.Println()
	//	}
	lat = locationmessages[0].Lat
	lon = locationmessages[0].Lon
	// dectect the kind of Language
	maps := map[string]string{"en": "en", "zh-CN": "zh_cn"}
	lang := "en"
	language, err := detectLanguage(city)
	if err != nil {
		log.Fatal(err)
	}
	tag := language.Language
	textString := tag.String()

	// fmt.Println(text)
	for k, v := range maps {
		if k == textString {
			lang = v
		}
	}
	url_weather := "https://api.openweathermap.org/data/2.5/weather?lat=" + strconv.FormatFloat(lat, 'f', -1, 32) + "&lon=" + strconv.FormatFloat(lon, 'f', -1, 32) + "&appid=" + api_key + "&lang=" + lang
	resp_w, err := http.Get(url_weather)
	if err != nil {
		log.Fatal(err)
	}
	defer resp_w.Body.Close()
	body_w, err := io.ReadAll(resp_w.Body)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(body_w, &weatherresponse); err != nil {
		log.Fatal(err)
	}
	for w := range weatherresponse.Weather {
		fmt.Printf("Main = %v, Description = %v", weatherresponse.Weather[w].Main, weatherresponse.Weather[w].Description)
		fmt.Println()
	}
	sayCurrentcity, _ := translateText(textString, "Current city is ")
	sayMain, _ := translateText(textString, "Main")
	sayDescription, _ := translateText(textString, "Description")
	sayWeatherMain, _ := translateText(textString, weatherresponse.Weather[0].Main)

	response := webhookResponse{
		FulfillmentMessages: []message{
			{
				Text: text{
					Text: []string{sayCurrentcity + city + sayMain + sayWeatherMain + sayDescription + weatherresponse.Weather[0].Description},
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
	case "weather":
		response, err = weatherQuery(request)
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
