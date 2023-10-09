package app

import (
	"fmt"
	"gitLab-rls-note/pkg/errors"
	"net/url"
)

const (
	mergeRequestState = "merged"
	issueState        = "closed"
)

type Service interface {
	RetrieveMergeRequests(query url.Values) ([]DecoratedMergeRequest, error)
	RetrieveIssues(query url.Values) ([]DecoratedIssue, error)
}
type service struct {
	client GitLabClient
}

func NewService(client GitLabClient) Service {
	return &service{client: client}
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

type GitLabClient interface {
	RetrieveIssues(query url.Values) ([]Issue, error)
	RetrieveRepo() (Repository, error)
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

type Repository struct {
}

type MergeRequest struct {
	IID    int      `json:"iid"`
	Title  string   `json:"title"`
	WebURL string   `json:"web_url"`
	Labels []string `json:"labels"`
	Author struct {
		Username string `json:"username"`
		WebURL   string `json:"web_url"`
	}
}

type Tag struct {
}

type CommitRef struct {
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
