package main

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/zalando/go-keyring"
)

const keyringService = "azure-devops-tui"
const keyringUser = "default"

// GetPAT retrieves the PAT from the OS keyring
func GetPAT() (string, error) {
	pat, err := keyring.Get(keyringService, keyringUser)
	if err != nil {
		return "", err
	}
	return pat, nil
}

// SetPAT stores the PAT in the OS keyring
func SetPAT(pat string) error {
	return keyring.Set(keyringService, keyringUser, pat)
}

// PromptPAT interactively prompts the user for their PAT (masked input)
func PromptPAT() (string, error) {
	var pat string
	prompt := &survey.Password{Message: "Enter your Azure DevOps Personal Access Token (PAT):"}
	err := survey.AskOne(prompt, &pat)
	if err != nil {
		return "", err
	}
	pat = strings.TrimSpace(pat)
	if pat == "" {
		return "", fmt.Errorf("PAT cannot be empty")
	}
	return pat, nil
}

// EnsurePAT checks for a stored PAT, prompts if missing, and stores securely
func EnsurePAT() (string, error) {
	pat, err := GetPAT()
	if err == nil && pat != "" {
		return pat, nil
	}
	pat, err = PromptPAT()
	if err != nil {
		return "", err
	}
	if err := SetPAT(pat); err != nil {
		return "", err
	}
	return pat, nil
}

// DeletePAT deletes the PAT from the keyring
func DeletePAT() error {
	return keyring.Delete(keyringService, keyringUser)
}
