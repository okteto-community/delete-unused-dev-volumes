package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const (
	// namespacesAPIPath is the path to the namespaces endpoint
	namespacesAPIPath = "/api/v0/namespaces"
)

func sendRequest(url, token string, response interface{}, logger *slog.Logger) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error("Error creating request")
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	// Create an HTTP client and send the request
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error sending request")
		return err
	}
	defer resp.Body.Close()

	// Check if the HTTP status is OK (200)
	if resp.StatusCode != http.StatusOK {
		logger.Error(fmt.Sprintf("Request failed. HTTP status code: %d", resp.StatusCode))
		return err
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&response)
	if err != nil {
		logger.Error("Error decoding response")
		return err
	}

	return nil
}
