package main

import (
	"net/url"

	"github.com/google/go-github/github"
)

type createMigrationRequest struct {
	AuthPassword string `json:"auth_password"`
	AuthUsername string `json:"auth_username"`
	CloneAddr    string `json:"clone_addr"`
	Description  string `json:"description"`
	Issues       bool   `json:"issues"`
	//Labels       bool   `json:"labels"`
	//Milestones   bool   `json:"milestones"`
	Mirror       bool `json:"mirror"`
	Private      bool `json:"private"`
	PullRequests bool `json:"pull_requests"`
	//Releases     bool   `json:"releases"`
	RepoName string `json:"repo_name"`
	Uid      int64  `json:"uid"`
	Wiki     bool   `json:"wiki"`
}

func createMigrationRequestFromGithubRepo(gr *github.Repository) createMigrationRequest {
	cmr := createMigrationRequest{
		CloneAddr:    strFromPtr(gr.CloneURL),
		Description:  strFromPtr(gr.Description),
		Issues:       boolFromPtr(gr.HasIssues),
		Mirror:       true,
		Private:      boolFromPtr(gr.Private),
		PullRequests: boolFromPtr(gr.HasIssues),
		RepoName:     strFromPtr(gr.Name),
		Uid:          cfg.TargetUser,
		Wiki:         boolFromPtr(gr.HasWiki),
	}

	if boolFromPtr(gr.Private) {
		uri, _ := url.Parse(strFromPtr(gr.CloneURL))
		uri.User = url.UserPassword("api", cfg.GithubToken)
		cmr.CloneAddr = uri.String()
	}

	return cmr
}

func boolFromPtr(in *bool) bool {
	return in != nil && *in
}

func strFromPtr(in *string) string {
	if in == nil {
		return ""
	}

	return *in
}
