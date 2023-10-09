package main

import (
	"fmt"
	"net/url"

	"gitLab-rls-note/app"
	"gitLab-rls-note/pkg/config"
	"gitLab-rls-note/store"
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

	client := store.NewGitlabClient(
		env.PersonalToken,
		env.APIEndpoint,
		env.ProjectID,
	)
	svc := app.NewService(client)

	mrs, err := svc.RetrieveMergeRequests(url.Values{
		"updated_before": {"2023-10-09T07:22:19.000+00:00"},
		"updated_after":  {"2023-10-09T07:15:26.000+00:00"},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(mrs)

	issues, err := svc.RetrieveIssues(url.Values{
		"updated_before": {"2023-10-09T07:22:19.000+00:00"},
		"updated_after":  {"2023-10-09T07:15:26.000+00:00"},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(issues)
}
