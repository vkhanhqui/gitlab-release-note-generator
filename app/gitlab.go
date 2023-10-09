package app

import (
	"fmt"
	"gitLab-rls-note/pkg/errors"
	"net/url"
	"regexp"
)

const (
	mergeRequestState = "merged"
	issueState        = "closed"
)

type Service interface {
	RetrieveMergeRequests(query url.Values) ([]DecoratedMergeRequest, error)
	RetrieveIssues(query url.Values) ([]DecoratedIssue, error)
	RetrieveTwoLatestTags() ([]Tag, error)
}
type service struct {
	client GitLabClient
	config Config
}

func NewService(client GitLabClient, config Config) Service {
	return &service{client: client, config: config}
}

func (s *service) RetrieveMergeRequests(query url.Values) ([]DecoratedMergeRequest, error) {
	query.Add("state", mergeRequestState)
	mrs, err := s.client.RetrieveMergeRequests(query)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var list []DecoratedMergeRequest
	for _, mr := range mrs {
		list = append(list, s.decorateMergeRequest(mr))
	}
	return list, nil
}

func (s *service) RetrieveIssues(query url.Values) ([]DecoratedIssue, error) {
	query.Add("state", issueState)
	issues, err := s.client.RetrieveIssues(query)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var list []DecoratedIssue
	for _, issue := range issues {
		list = append(list, s.decorateIssue(issue))
	}
	return list, nil
}

func (s *service) RetrieveTwoLatestTags() ([]Tag, error) {
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
	return []Tag{latest, *secondTag}, nil
}

func (s *service) decorateMergeRequest(mr MergeRequest) DecoratedMergeRequest {
	msg := fmt.Sprintf("- %s [#%d](%s) ([%s](%s))", mr.Title, mr.IID, mr.WebURL, mr.Author.Username, mr.Author.WebURL)
	return DecoratedMergeRequest{
		Message:      msg,
		Labels:       mr.Labels,
		DefaultLabel: "mergeRequests",
	}
}

func (s *service) decorateIssue(issue Issue) DecoratedIssue {
	msg := fmt.Sprintf("- %s [#%d](%s)", issue.Title, issue.IID, issue.WebURL)
	return DecoratedIssue{
		Message:      msg,
		Labels:       issue.Labels,
		DefaultLabel: "issues",
	}
}

func (s *service) isMatchTargetTagRegex(tag Tag) (bool, error) {
	regex, err := regexp.Compile(s.config.TargetTagRegex)
	if err != nil {
		return false, errors.WithStack(err)
	}
	return regex.MatchString(tag.Name), nil
}

func (s *service) isInTargetBranch(commits []CommitRef) bool {
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
	CreateTagRelease(tagName string, body Release) error
	UpdateTagRelease(tagName string, body Release) error
}

type Issue struct {
	IID    int      `json:"iid"`
	Title  string   `json:"title"`
	WebURL string   `json:"web_url"`
	Labels []string `json:"labels"`
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
}

type Tag struct {
	Name   string `json:"name"`
	Commit Commit `json:"commit"`
}

type Commit struct {
	ID            string `json:"id"`
	CommittedDate string `json:"committed_date"`
}

type CommitRef struct {
	Name string `json:"name"`
}

type Release struct {
}

type DecoratedMergeRequest struct {
	Message      string
	Labels       []string
	DefaultLabel string
}

type DecoratedIssue struct {
	Message      string
	Labels       []string
	DefaultLabel string
}

type Config struct {
	TargetBranch       string
	TargetTagRegex     string
	TZ                 string
	IssueClosedSeconds int
}
