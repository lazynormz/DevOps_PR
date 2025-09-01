package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
)

type PullrequestReviewer struct {
	id          string
	displayName string
	isRequired  bool
	vote        int
}

type PullRequestInfo struct {
	id        int
	title     string
	creator   string
	creatorID string
	IsDraft   bool
	reviewers []PullrequestReviewer
}

func main() {
	args := os.Args
	if len(args) > 1 && args[1] == "reset" {
		if err := DeletePAT(); err != nil {
			fmt.Println("Error deleting PAT:", err)
		}
		fmt.Println("PAT has been reset. Please enter a new PAT when prompted.")
	}
	baseAddress := "https://dev.azure.com"
	organization := "2care4"
	project := "IT2care4"

	PAT, err := EnsurePAT()
	if err != nil {
		fmt.Println("Error retrieving PAT:", err)
		return
	}

	fullUrl := baseAddress + "/" + organization + "/"

	ctx := context.Background()

	tryCount := 0
	maxTries := 2

	for tryCount < maxTries {
		connection := azuredevops.NewPatConnection(fullUrl, PAT)
		gitClient, err := git.NewClient(ctx, connection)
		if err != nil {
			if strings.Contains(err.Error(), "401") {
				fmt.Println("PAT is invalid or expired. Please enter a new PAT.")
				PAT, err = PromptPAT()
				if err != nil {
					fmt.Println("Error reading PAT:", err)
					return
				}
				if err := SetPAT(PAT); err != nil {
					fmt.Println("Error saving PAT:", err)
					return
				}
				tryCount++
				continue
			}
			panic(err)
		}

		pullRequests, err := ListOpenPullRequests(ctx, gitClient, project)
		if err != nil {
			if strings.Contains(err.Error(), "401") {
				fmt.Println("PAT is invalid or expired. Please enter a new PAT.")
				PAT, err = PromptPAT()
				if err != nil {
					fmt.Println("Error reading PAT:", err)
					return
				}
				if err := SetPAT(PAT); err != nil {
					fmt.Println("Error saving PAT:", err)
					return
				}
				tryCount++
				continue
			}
			panic(err)
		}
		userID, err := GetCurrentUserID(PAT)
		if err != nil {
			// Instead of panicking, launch TUI with error message
			RunTUIWithError(pullRequests, err.Error())
			return
		}
		RunTUI(pullRequests, userID)
		return
	}
}

// printPullRequest prints details of a pull request
func printPullRequest(pr PullRequestInfo) {
	println("PR ID:", pr.id)
	println("Title:", pr.title)
	println("Creator:", pr.creator)
	println("Reviewers:")
	for _, rev := range pr.reviewers {
		println(" -", rev.displayName, "(ID:", rev.id, "Required:", rev.isRequired, "Vote:", rev.vote, ")")
	}
	println("-----")
}

// ListOpenPullRequests lists all open pull requests in all repositories in the specified project
func ListOpenPullRequests(ctx context.Context, gitClient git.Client, project string) ([]PullRequestInfo, error) {
	repos, err := gitClient.GetRepositories(ctx, git.GetRepositoriesArgs{
		Project: &project,
	})
	if err != nil {
		return nil, err
	}
	if repos == nil {
		return nil, nil
	}
	var allPRs []PullRequestInfo
	statusActive := git.PullRequestStatus("active")
	for _, repo := range *repos {
		repoIdStr := repo.Id.String()
		prs, err := gitClient.GetPullRequests(ctx, git.GetPullRequestsArgs{
			RepositoryId: &repoIdStr,
			SearchCriteria: &git.GitPullRequestSearchCriteria{
				Status: &statusActive,
			},
			Project: &project,
		})
		if err != nil {
			return nil, err
		}
		if prs != nil {
			for _, pr := range *prs {
				allPRs = append(allPRs, createPullRequestInfo(&pr))
			}
		}
	}
	return allPRs, nil
}
