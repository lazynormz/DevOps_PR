package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// GetCurrentUserID fetches the current user's Azure DevOps ID using the PAT
func GetCurrentUserID(pat string) (string, error) {
	organization := "2care4" // TODO: optionally pass this as a parameter
	url := fmt.Sprintf("https://vssps.dev.azure.com/%s/_apis/profile/profiles/me?api-version=7.0", organization)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth("", pat)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch user profile: status %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
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
