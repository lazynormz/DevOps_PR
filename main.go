package main

import (
	"context"

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
	IsDraft   bool
	reviewers []PullrequestReviewer
}

func main() {
	baseAddress := "https://dev.azure.com"
	organization := "2care4"
	project := "IT2care4"

	PAT := "EwqSbpHwptodPbag810103rVwkeRRBgsrZwn41cGQ5cz7F13mvnSJQQJ99BHACAAAAAsNLvJAAASAZDO2bdb"

	fullUrl := baseAddress + "/" + organization + "/"

	connection := azuredevops.NewPatConnection(fullUrl, PAT)

	ctx := context.Background()

	gitClient, err := git.NewClient(ctx, connection)
	if err != nil {
		panic(err)
	}

	pullRequests, err := ListOpenPullRequests(ctx, gitClient, project)
	if err != nil {
		panic(err)
	}

	for _, pr := range pullRequests {
		if pr.IsDraft {
			continue
		}

		printPullRequest(pr)
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
