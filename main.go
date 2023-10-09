package main

import (
	"fmt"

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
	svc := app.NewService(client, app.Config{
		TargetBranch:       env.TargetBranch,
		TargetTagRegex:     env.TargetTagRegex,
		TZ:                 env.TZ,
		IssueClosedSeconds: env.IssueClosedSeconds,
	})

	tags, err := svc.RetrieveTwoLatestTags()
	if err != nil {
		panic(err)
	}
	fmt.Println(tags)
}
