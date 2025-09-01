package main

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/zalando/go-keyring"
)

func DeleteOrganization() error {
	return keyring.Delete(orgKeyringService, orgKeyringUser)
}

func DeleteProject() error {
	return keyring.Delete(projKeyringService, projKeyringUser)
}

const orgKeyringService = "azure-devops-tui-org"
const orgKeyringUser = "default"
const projKeyringService = "azure-devops-tui-project"
const projKeyringUser = "default"

func GetOrganization() (string, error) {
	org, err := keyring.Get(orgKeyringService, orgKeyringUser)
	if err != nil {
		return "", err
	}
	return org, nil
}

func SetOrganization(org string) error {
	return keyring.Set(orgKeyringService, orgKeyringUser, org)
}

func PromptOrganization() (string, error) {
	var org string
	prompt := &survey.Input{Message: "Enter your Azure DevOps organization:"}
	err := survey.AskOne(prompt, &org)
	if err != nil {
		return "", err
	}
	org = strings.TrimSpace(org)
	if org == "" {
		return "", fmt.Errorf("organization cannot be empty")
	}
	return org, nil
}

func EnsureOrganization() (string, error) {
	org, err := GetOrganization()
	if err == nil && org != "" {
		return org, nil
	}
	org, err = PromptOrganization()
	if err != nil {
		return "", err
	}
	if err := SetOrganization(org); err != nil {
		return "", err
	}
	return org, nil
}

func GetProject() (string, error) {
	proj, err := keyring.Get(projKeyringService, projKeyringUser)
	if err != nil {
		return "", err
	}
	return proj, nil
}

func SetProject(proj string) error {
	return keyring.Set(projKeyringService, projKeyringUser, proj)
}

func PromptProject() (string, error) {
	var proj string
	prompt := &survey.Input{Message: "Enter your Azure DevOps project:"}
	err := survey.AskOne(prompt, &proj)
	if err != nil {
		return "", err
	}
	proj = strings.TrimSpace(proj)
	if proj == "" {
		return "", fmt.Errorf("project cannot be empty")
	}
	return proj, nil
}

func EnsureProject() (string, error) {
	proj, err := GetProject()
	if err == nil && proj != "" {
		return proj, nil
	}
	proj, err = PromptProject()
	if err != nil {
		return "", err
	}
	if err := SetProject(proj); err != nil {
		return "", err
	}
	return proj, nil
}
