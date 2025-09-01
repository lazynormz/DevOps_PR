package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GetCurrentUserID fetches the current user's Azure DevOps ID using the PAT
func GetCurrentUserID(pat string, organization string) (string, error) {
	url := fmt.Sprintf("https://vssps.dev.azure.com/%s/_apis/profile/profiles/me?api-version=7.0", organization)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	request.SetBasicAuth("", pat)
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch user profile: status %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	if len(body) == 0 {
		return "", fmt.Errorf("empty response from user profile API")
	}

	var result struct {
		Id string `json:"id"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse user profile JSON: %w", err)
	}
	if result.Id == "" {
		return "", fmt.Errorf("user ID not found in profile response")
	}
	return result.Id, nil
}
