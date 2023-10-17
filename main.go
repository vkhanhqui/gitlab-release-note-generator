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
	TimeZone           string `mapstructure:"TZ"`
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
	svc := app.NewGitLabService(client, app.Config{
		TargetBranch:   env.TargetBranch,
		TargetTagRegex: env.TargetTagRegex,
		TimeZone:       env.TimeZone,
	})

	tags, err := svc.RetrieveTwoLatestTags()
	if err != nil {
		panic(err)
	}
	fmt.Println(tags)

	if len(tags) != 2 {
		fmt.Println("Cannot find latest and second latest tag. Abort the program!")
	}

	latestTag, secondLatestTag := tags[0], tags[1]
	if latestTag.Commit.CommittedDate == "" || secondLatestTag.Commit.CommittedDate == "" {
		fmt.Println("Cannot find latest and second latest tag. Abort the program!")
	}

	startDate := secondLatestTag.Commit.CommittedDate
	endDate := latestTag.Commit.CommittedDate
	// if env.IssueClosedSeconds > 0 {

	// }
	mrs, issues, err := svc.RetrieveChangelogsByStartAndEndDate(startDate, endDate)
	if err != nil {
		panic(err)
	}
	fmt.Println(mrs)
	fmt.Println(issues)

}
