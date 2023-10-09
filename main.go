package main

import (
	"fmt"
	"net/url"

	"gitlab-release-note-generator/app"
	"gitlab-release-note-generator/pkg/config"
	"gitlab-release-note-generator/store"
)

type envConfig struct {
	PersonalToken      string `mapstructure:"GITLAB_PERSONAL_TOKEN"`
	APIEndpoint        string `mapstructure:"GITLAB_API_ENDPOINT"`
	ProjectID          string `mapstructure:"GITLAB_PROJECT_ID"`
	TargetBranch       string `mapstructure:"TARGET_BRANCH"`
	TargetTagRegex     string `mapstructure:"TARGET_TAG_REGEX"`
	TZ                 string `mapstructure:"TZ"`
	IssueClosedSeconds int    `mapstructure:"ISSUE_CLOSED_SECONDS"`
}

func main() {
	config.LoadEnvConfig()
	env := envConfig{}
	err := config.UnmarshalEnvConfig(&env)
	if err != nil {
		panic(err)
	}

	var gitlabAdapter app.GitLabClient = store.NewGitlabClient(
		env.PersonalToken,
		env.APIEndpoint,
		env.ProjectID,
	)

	query := url.Values{
		"state":          {"merged"},
		"updated_before": {"2023-10-09T07:22:19.000+00:00"},
		"updated_after":  {"2023-10-09T07:15:26.000+00:00"},
	}
	mrs, err := gitlabAdapter.SearchMergeRequests(query)
	if err != nil {
		panic(err)
	}
	fmt.Println(app.DecorateMergeRequests(mrs))
}
