package app

import (
	"gitLab-rls-note/pkg/errors"
	"net/url"
	"regexp"
	"time"
)

const (
	mergeRequestState = "merged"
	issueState        = "closed"
	gitlabTimeFormat  = time.RFC3339Nano
)

type GitLabService interface {
	RetrieveTwoLatestTags() ([]Tag, error)
	RetrieveChangelogsByStartAndEndDate(startDate, endDate string) ([]MergeRequest, []Issue, error)
	Publish(tag Tag, content string) error
}

type gitLabService struct {
	client GitLabClient
	config Config
}

type Config struct {
	TargetBranch       string
	TargetTagRegex     string
	IssueClosedSeconds int
}

func NewGitLabService(client GitLabClient, config Config) GitLabService {
	return &gitLabService{client: client, config: config}
}

func (s *gitLabService) Publish(tag Tag, content string) error {
	body := Release{tag.Name, content}
	if tag.Release.Name != "" {
		err := s.client.UpdateTagRelease(body)
		return err
	}
	err := s.client.CreateTagRelease(body)
	return err
}

func (s *gitLabService) RetrieveChangelogsByStartAndEndDate(startDate, endDate string) ([]MergeRequest, []Issue, error) {
	mergeRequests, err := s.retrieveMergeRequests(url.Values{
		"updated_before": {endDate},
		"updated_after":  {startDate},
	})
	if err != nil {
		return nil, nil, err
	}

	parsedStartDate, err := time.Parse(gitlabTimeFormat, startDate)
	if err != nil {
		return nil, nil, err
	}

	parsedEndDate, err := time.Parse(gitlabTimeFormat, endDate)
	if err != nil {
		return nil, nil, err
	}

	var filteredMRs []MergeRequest
	for _, mr := range mergeRequests {
		parsedTime, err := time.Parse(gitlabTimeFormat, mr.MergedAt)
		if err != nil {
			return nil, nil, err
		}

		if parsedTime.After(parsedStartDate) && parsedTime.Before(parsedEndDate) {
			filteredMRs = append(filteredMRs, mr)
		}
	}

	issues, err := s.retrieveIssues(url.Values{
		"updated_before": {endDate},
		"updated_after":  {startDate},
	})
	if err != nil {
		return nil, nil, err
	}

	var filteredISs []Issue
	for _, iss := range issues {
		parsedTime, err := time.Parse(gitlabTimeFormat, iss.ClosedAt)
		if err != nil {
			return nil, nil, err
		}

		if parsedTime.After(parsedStartDate) && parsedTime.Before(parsedEndDate) {
			filteredISs = append(filteredISs, iss)
		}
	}

	return filteredMRs, filteredISs, nil
}

func (s *gitLabService) RetrieveTwoLatestTags() ([]Tag, error) {
	tags, err := s.client.RetrieveTags(url.Values{})
	if err != nil || len(tags) < 1 {
		return nil, err
	}

	latest := tags[0]
	isMatch, err := s.isMatchTargetTagRegex(latest)
	if err != nil || !isMatch {
		return nil, err
	}

	latestCommits, err := s.client.RetrieveCommitRefsBySHA(latest.Commit.ID, url.Values{"type": {"branch"}})
	if err != nil {
		return nil, err
	}

	if !s.isInTargetBranch(latestCommits) {
		return nil, errors.New("Latest tag doesn't contain target branch.")
	}

	tags = tags[1:]
	if len(tags) == 0 {
		repo, err := s.client.RetrieveRepo()
		if err != nil {
			return nil, err
		}

		return []Tag{latest, {
			Commit: Commit{
				CommittedDate: repo.CreatedAt,
			},
		}}, nil
	}

	var secondTag *Tag
	for _, tag := range tags {
		isMatch, err := s.isMatchTargetTagRegex(tag)
		if err != nil || !isMatch {
			return nil, err
		}

		commits, err := s.client.RetrieveCommitRefsBySHA(tag.Commit.ID, url.Values{"type": {"branch"}})
		if err != nil {
			return nil, err
		}

		if s.isInTargetBranch(commits) {
			secondTag = &tag
			break
		}
	}

	if s.config.IssueClosedSeconds > 0 {
		parsedReleaseDate, err := time.Parse(gitlabTimeFormat, latest.Commit.CommittedDate)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		latest.Commit.CommittedDate = parsedReleaseDate.Add(time.Duration(s.config.IssueClosedSeconds) * time.Second).Format(gitlabTimeFormat)
	}
	return []Tag{latest, *secondTag}, nil
}

func (s *gitLabService) retrieveMergeRequests(query url.Values) ([]MergeRequest, error) {
	query.Add("state", mergeRequestState)
	list, err := s.client.RetrieveMergeRequests(query)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return list, nil
}

func (s *gitLabService) retrieveIssues(query url.Values) ([]Issue, error) {
	query.Add("state", issueState)
	list, err := s.client.RetrieveIssues(query)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return list, nil
}

func (s *gitLabService) isMatchTargetTagRegex(tag Tag) (bool, error) {
	regex, err := regexp.Compile(s.config.TargetTagRegex)
	if err != nil {
		return false, errors.WithStack(err)
	}
	return regex.MatchString(tag.Name), nil
}

func (s *gitLabService) isInTargetBranch(commits []CommitRef) bool {
	if s.config.TargetBranch == "" {
		return true
	}
	for _, commit := range commits {
		if commit.Name == s.config.TargetBranch {
			return true
		}
	}
	return false
}

type GitLabClient interface {
	RetrieveIssues(query url.Values) ([]Issue, error)
	RetrieveRepo() (Repo, error)
	RetrieveMergeRequests(query url.Values) ([]MergeRequest, error)
	RetrieveTags(query url.Values) ([]Tag, error)
	RetrieveCommitRefsBySHA(sha string, query url.Values) ([]CommitRef, error)
	CreateTagRelease(body Release) error
	UpdateTagRelease(body Release) error
}

type Issue struct {
	IID      int      `json:"iid"`
	Title    string   `json:"title"`
	WebURL   string   `json:"web_url"`
	Labels   []string `json:"labels"`
	ClosedAt string   `json:"closed_at"`
}

type Repo struct {
	CreatedAt string `json:"created_at"`
}

type MergeRequest struct {
	IID    int      `json:"iid"`
	Title  string   `json:"title"`
	WebURL string   `json:"web_url"`
	Labels []string `json:"labels"`
	Author struct {
		Username string `json:"username"`
		WebURL   string `json:"web_url"`
	} `json:"author"`
	MergedAt string `json:"merged_at"`
}

type Tag struct {
	Name    string  `json:"name"`
	Commit  Commit  `json:"commit"`
	Release Release `json:"release"`
}

type Commit struct {
	ID            string `json:"id"`
	CommittedDate string `json:"committed_date"`
}

type CommitRef struct {
	Name string `json:"name"`
}

type Release struct {
	Name        string `json:"tag_name"`
	Description string `json:"description"`
}
