package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type UserResponseBody struct {
	Id string `json:"id"`
}

type GetUserResponseBody struct {
	Data []UserResponseBody `json:"data"`
}

func GetUserId(token string) (int, error) {
	getUserRequest, err := http.NewRequest("GET", "https://api.twitch.tv/helix/users", nil)

	if err != nil {
		fmt.Println("Could not create Request for User ID", err)
		os.Exit(1)
	}

	getUserRequest.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	getUserRequest.Header.Add("Client-Id", CLIENT_ID)

	getUserResponse, err := httpClient.Do(getUserRequest)

	if err != nil {
		fmt.Println("Could not fetch user ID from Twitch", err)
		os.Exit(1)
	}

	if getUserResponse.StatusCode != 200 {
		bodyStr, _ := io.ReadAll(getUserResponse.Body)
		fmt.Println("Could not create user ID request", string(bodyStr))
		os.Exit(1)
	}

	var getUserBody GetUserResponseBody
	err = json.NewDecoder(getUserResponse.Body).Decode(&getUserBody)
	if err != nil {
		fmt.Println("Could not decode request body at POST:Auth")
		os.Exit(1)
	}

	if len(getUserBody.Data) == 0 {
		fmt.Println("Could not fetch user Id. data was empty")
		os.Exit(1)
	}

	return 0, err
}

type CreatePredictionResult struct {
	PredictionId     string
	SuccessOutcomeId string
	FailureOutcomeId string
}

type CreatePredictionOutcome struct {
	Id string `json:"id"`
}

type CreatePredictionData struct {
	Id               string                    `json:"id"`
	WinningOutcomeId string                    `json:"winning_outcome_id"`
	Outcomes         []CreatePredictionOutcome `json:"outcomes"`
	Status           string                    `json:"status"`
}

type CreatePredictionBody struct {
	Data []CreatePredictionData `json:"data"`
}

func CreatePrediction(token string, title string, userId string, successMessage string, failureMessage string, duration int) CreatePredictionResult {

	createPredictionRequestBodyStr := fmt.Sprintf(`{
		"broadcaster_id": "%s",
		"title": "%s",
		"outcomes": [
			{
				"title": "%s"
			},
			{
				"title": "%s"
			}
		],
		"prediction_window": %d
	}`, userId, title, successMessage, failureMessage, duration)

	createPredictionRequestBody := []byte(createPredictionRequestBodyStr)

	createPredictionRequest, err := http.NewRequest("POST", "https://api.twitch.tv/helix/predictions", bytes.NewBuffer(createPredictionRequestBody))
	if err != nil {
		fmt.Println("Could not create Request for Prediction", err)
		os.Exit(1)
	}

	createPredictionRequest.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	createPredictionRequest.Header.Add("Client-Id", CLIENT_ID)
	createPredictionRequest.Header.Add("Content-Type", "application/json")

	createPredictionResponse, err := httpClient.Do(createPredictionRequest)

	if err != nil {
		fmt.Println("Could not create prediction", err)
		os.Exit(1)
	}

	if createPredictionResponse.StatusCode != 200 {
		bodyStr, _ := io.ReadAll(createPredictionResponse.Body)
		fmt.Println("Could not create prediction", string(bodyStr))
		os.Exit(1)
	}

	var createPredictionResponseBody CreatePredictionBody
	err = json.NewDecoder(createPredictionResponse.Body).Decode(&createPredictionResponseBody)
	if err != nil {
		fmt.Println("Could not decode request body at create prediction", err)
		os.Exit(1)
	}

	if len(createPredictionResponseBody.Data) == 0 {
		fmt.Println("Could not create prediction. data was empty")
		os.Exit(1)
	}

	prediction := createPredictionResponseBody.Data[0]

	return CreatePredictionResult{
		PredictionId:     prediction.Id,
		SuccessOutcomeId: prediction.Outcomes[0].Id,
		FailureOutcomeId: prediction.Outcomes[1].Id,
	}

}

func ResolvePrediction(token string, userId string, predictionId string, winningOutcomeId string) {

	resolvePredictionRequestBody := []byte(fmt.Sprintf(`{
		"broadcaster_id": "%s",
		"winning_outcome_id": "%s",
		"id": "%s",
		"status": "RESOLVED"
	}`, userId, winningOutcomeId, predictionId))

	resolvePredictionRequest, err := http.NewRequest("PATCH", "https://api.twitch.tv/helix/predictions", bytes.NewBuffer(resolvePredictionRequestBody))
	if err != nil {
		fmt.Println("Could not resolve Request for User ID", err)
		os.Exit(1)
	}

	resolvePredictionRequest.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	resolvePredictionRequest.Header.Add("Client-Id", CLIENT_ID)
	resolvePredictionRequest.Header.Add("Content-Type", "application/json")

	resolvePredictionResponse, err := httpClient.Do(resolvePredictionRequest)

	if resolvePredictionResponse.StatusCode != 200 {
		bodyStr, _ := io.ReadAll(resolvePredictionResponse.Body)
		fmt.Println("Could not resolve prediction", string(bodyStr))
		os.Exit(1)
	}

	if err != nil {
		fmt.Println("Could not resolve prediction", err)
		os.Exit(1)
	}
}

func IsPredictionFinished(token string, userId string, predictionId string) bool {
	getPredictionRequest, err := http.NewRequest("GET", fmt.Sprintf("https://api.twitch.tv/helix/predictions?id=%s&broadcaster_id=%s", predictionId, userId), nil)

	if err != nil {
		fmt.Println("Could not create request for getting existing prediction", err)
		os.Exit(1)
	}

	getPredictionRequest.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	getPredictionRequest.Header.Add("Client-Id", CLIENT_ID)

	getPredictionResponse, err := httpClient.Do(getPredictionRequest)

	if err != nil {
		fmt.Println("Could not get prediction", predictionId, err)
		os.Exit(1)
	}

	if getPredictionResponse.StatusCode != 200 {
		bodyStr, _ := io.ReadAll(getPredictionResponse.Body)
		fmt.Println("Could not get prediction", string(bodyStr))
		os.Exit(1)
	}

	var getPredictionResponseBody CreatePredictionBody
	err = json.NewDecoder(getPredictionResponse.Body).Decode(&getPredictionResponseBody)
	if err != nil {
		fmt.Println("Could not decode request body at get prediction")
		os.Exit(1)
	}

	if len(getPredictionResponseBody.Data) == 0 {
		fmt.Println("Could not get prediction. data was empty")
		os.Exit(1)
	}

	prediction := getPredictionResponseBody.Data[0]

	if prediction.Status == "LOCKED" {
		return true
	} else {
		return false
	}

}
