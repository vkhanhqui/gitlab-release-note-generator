package app

import (
	"fmt"
	"gitLab-rls-note/pkg/errors"
	"net/url"
)

type ContentService interface {
	RetrieveMergeRequests(query url.Values) ([]DecoratedMergeRequest, error)
	RetrieveIssues(query url.Values) ([]DecoratedIssue, error)
}
type contentService struct {
	client GitLabClient
	config Config
}

func NewContentService(client GitLabClient, config Config) ContentService {
	return &contentService{client: client, config: config}
}

func (s *contentService) RetrieveMergeRequests(query url.Values) ([]DecoratedMergeRequest, error) {
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

func (s *contentService) RetrieveIssues(query url.Values) ([]DecoratedIssue, error) {
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

func (s *contentService) decorateMergeRequest(mr MergeRequest) DecoratedMergeRequest {
	msg := fmt.Sprintf("- %s [#%d](%s) ([%s](%s))", mr.Title, mr.IID, mr.WebURL, mr.Author.Username, mr.Author.WebURL)
	return DecoratedMergeRequest{
		Message:      msg,
		Labels:       mr.Labels,
		DefaultLabel: "mergeRequests",
	}
}

func (s *contentService) decorateIssue(issue Issue) DecoratedIssue {
	msg := fmt.Sprintf("- %s [#%d](%s)", issue.Title, issue.IID, issue.WebURL)
	return DecoratedIssue{
		Message:      msg,
		Labels:       issue.Labels,
		DefaultLabel: "issues",
	}
}
