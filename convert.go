package main

import "github.com/microsoft/azure-devops-go-api/azuredevops/git"

// createPullRequestInfo converts a GitPullRequest to a PullRequestInfo struct
func createPullRequestInfo(pr *git.GitPullRequest) PullRequestInfo {
	var reviewers []PullrequestReviewer
	if pr.Reviewers != nil {
		for _, rev := range *pr.Reviewers {
			reviewer := PullrequestReviewer{
				id:          derefString(rev.Id),
				displayName: derefString(rev.DisplayName),
				isRequired:  derefBool(rev.IsRequired),
				vote:        derefInt(rev.Vote),
			}
			reviewers = append(reviewers, reviewer)
		}
	}
	return PullRequestInfo{
		id:        derefInt(pr.PullRequestId),
		title:     derefString(pr.Title),
		creator:   derefString(pr.CreatedBy.DisplayName),
		IsDraft:   derefBool(pr.IsDraft),
		reviewers: reviewers,
	}
}
