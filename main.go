package main

import (
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
	TimeZone           string `mapstructure:"TZ"`
	IssueClosedSeconds int    `mapstructure:"ISSUE_CLOSED_SECONDS"`
	ZeroTrustCookie    string `mapstructure:"ZERO_TRUST_COOKIE"`
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
		env.ZeroTrustCookie,
	)
	gitLabSvc := app.NewGitLabService(client, app.Config{
		TargetBranch:       env.TargetBranch,
		TargetTagRegex:     env.TargetTagRegex,
		IssueClosedSeconds: env.IssueClosedSeconds,
	})

	tags, err := gitLabSvc.RetrieveTwoLatestTags()
	if err != nil {
		panic(err)
	}

	latestTag, secondLatestTag := tags[0], tags[1]
	startDate := secondLatestTag.Commit.CommittedDate
	endDate := latestTag.Commit.CommittedDate

	mrs, issues, err := gitLabSvc.RetrieveChangelogsByStartAndEndDate(startDate, endDate)
	if err != nil {
		panic(err)
	}

	contentSvc, err := app.NewContentService(env.TimeZone)
	if err != nil {
		panic(err)
	}
	content, err := contentSvc.GenerateContent(mrs, issues, endDate)
	if err != nil {
		panic(err)
	}

	err = gitLabSvc.Publish(latestTag, content)
	if err != nil {
		panic(err)
	}
}
