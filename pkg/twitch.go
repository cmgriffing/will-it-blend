package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Constants
const (
    TwitchAPIURL = "https://api.twitch.tv/helix"
)

// UserResponseBody represents the response body for a user request
type UserResponseBody struct {
    Id string `json:"id"`
}

// GetUserResponseBody represents the response body for getting user information
type GetUserResponseBody struct {
    Data []UserResponseBody `json:"data"`
}

// GetUserId fetches the user ID from Twitch
func GetUserId(token string) (string, error) {
    request, err := http.NewRequest("GET", TwitchAPIURL+"/users", nil)
    if err != nil {
        return "", fmt.Errorf("failed to create request for User ID: %w", err)
    }

    request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
    request.Header.Add("Client-Id", CLIENT_ID)

    response, err := httpClient.Do(request)
    if err != nil {
        return "", fmt.Errorf("failed to fetch user ID from Twitch: %w", err)
    }
    defer response.Body.Close()

	if response.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(response.Body)
		bodyStr := string(bodyBytes)
		return "", fmt.Errorf("failed to fetch user ID: %s", bodyStr)
	}

    var getUserBody GetUserResponseBody
    err = json.NewDecoder(response.Body).Decode(&getUserBody)
    if err != nil {
        return "", fmt.Errorf("failed to decode response body: %w", err)
    }

    if len(getUserBody.Data) == 0 {
        return "", errors.New("user data was empty")
    }

    return getUserBody.Data[0].Id, nil
}

// CreatePredictionResult represents the result of a prediction creation
type CreatePredictionResult struct {
    PredictionId     string
    SuccessOutcomeId string
    FailureOutcomeId string
}

// CreatePredictionOutcome represents the outcome structure in a prediction
type CreatePredictionOutcome struct {
    Id string `json:"id"`
}

// CreatePredictionData represents the prediction data structure
type CreatePredictionData struct {
    Id               string                    `json:"id"`
    WinningOutcomeId string                    `json:"winning_outcome_id"`
    Outcomes         []CreatePredictionOutcome `json:"outcomes"`
    Status           string                    `json:"status"`
}

// CreatePredictionBody represents the body of a prediction creation response
type CreatePredictionBody struct {
    Data []CreatePredictionData `json:"data"`
}

// CreatePrediction creates a prediction on Twitch
func CreatePrediction(token string, title string, userId string, successMessage string, failureMessage string, duration int) (CreatePredictionResult, error) {
    requestBody := fmt.Sprintf(`{
        "broadcaster_id": "%s",
        "title": "%s",
        "outcomes": [
            {"title": "%s"},
            {"title": "%s"}
        ],
        "prediction_window": %d
    }`, userId, title, successMessage, failureMessage, duration)

    request, err := http.NewRequest("POST", TwitchAPIURL+"/predictions", bytes.NewBuffer([]byte(requestBody)))
    if err != nil {
        return CreatePredictionResult{}, fmt.Errorf("failed to create request for prediction: %w", err)
    }

    addRequestHeaders(request, token)

    response, err := httpClient.Do(request)
    if err != nil {
        return CreatePredictionResult{}, fmt.Errorf("failed to create prediction: %w", err)
    }
    defer response.Body.Close()

    if response.StatusCode != 200 {
        bodyStr, _ := io.ReadAll(response.Body)
        return CreatePredictionResult{}, fmt.Errorf("failed to create prediction: %s", bodyStr)
    }

    var responseBody CreatePredictionBody
    err = json.NewDecoder(response.Body).Decode(&responseBody)
    if err != nil {
        return CreatePredictionResult{}, fmt.Errorf("failed to decode response body: %w", err)
    }

    if len(responseBody.Data) == 0 {
        return CreatePredictionResult{}, errors.New("prediction data was empty")
    }

    prediction := responseBody.Data[0]
    return CreatePredictionResult{
        PredictionId:     prediction.Id,
        SuccessOutcomeId: prediction.Outcomes[0].Id,
        FailureOutcomeId: prediction.Outcomes[1].Id,
    }, nil
}

// ResolvePrediction resolves a Twitch prediction
func ResolvePrediction(token string, userId string, predictionId string, winningOutcomeId string) error {
    requestBody := fmt.Sprintf(`{
        "broadcaster_id": "%s",
        "winning_outcome_id": "%s",
        "id": "%s",
        "status": "RESOLVED"
    }`, userId, winningOutcomeId, predictionId)

    request, err := http.NewRequest("PATCH", TwitchAPIURL+"/predictions", bytes.NewBuffer([]byte(requestBody)))
    if err != nil {
        return fmt.Errorf("failed to create request for resolving prediction: %w", err)
    }

    addRequestHeaders(request, token)

    response, err := httpClient.Do(request)
    if err != nil {
        return fmt.Errorf("failed to resolve prediction: %w", err)
    }
    defer response.Body.Close()

    if response.StatusCode != 200 {
        bodyStr, _ := io.ReadAll(response.Body)
        return fmt.Errorf("failed to resolve prediction: %s", bodyStr)
    }

    return nil
}

// IsPredictionFinished checks if a Twitch prediction is finished
func IsPredictionFinished(token string, userId string, predictionId string) (bool, error) {
    request, err := http.NewRequest("GET", fmt.Sprintf("%s/predictions?id=%s&broadcaster_id=%s", TwitchAPIURL, predictionId, userId), nil)
    if err != nil {
        return false, fmt.Errorf("failed to create request for checking prediction: %w", err)
    }

    addRequestHeaders(request, token)

    response, err := httpClient.Do(request)
    if err != nil {
        return false, fmt.Errorf("failed to check prediction: %w", err)
    }
    defer response.Body.Close()

    if response.StatusCode != 200 {
        bodyStr, _ := io.ReadAll(response.Body)
        return false, fmt.Errorf("failed to check prediction: %s", bodyStr)
    }

    var responseBody CreatePredictionBody
    err = json.NewDecoder(response.Body).Decode(&responseBody)
    if err != nil {
        return false, fmt.Errorf("failed to decode response body: %w", err)
    }

    if len(responseBody.Data) == 0 {
        return false, errors.New("prediction data was empty")
    }

    return responseBody.Data[0].Status == "LOCKED", nil
}

// addRequestHeaders adds common headers to an HTTP request
func addRequestHeaders(request *http.Request, token string) {
    request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
    request.Header.Add("Client-Id", CLIENT_ID)
    request.Header.Add("Content-Type", "application/json")
}
