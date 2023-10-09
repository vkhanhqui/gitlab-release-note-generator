package app

import (
	"fmt"
	"net/url"
)

func DecorateMergeRequests(mergeRequests []MergeRequest) []DecoratedMergeRequest {
	var list []DecoratedMergeRequest
	for _, mr := range mergeRequests {
		list = append(list, decorateMergeRequest(mr))
	}
	return list
}

func decorateMergeRequest(mr MergeRequest) DecoratedMergeRequest {
	msg := fmt.Sprintf("- %s [#%d](%s) ([%s](%s))", mr.Title, mr.IID, mr.WebURL, mr.Author.Username, mr.Author.WebURL)
	return DecoratedMergeRequest{
		Message:      msg,
		Labels:       mr.Labels,
		DefaultLabel: "mergeRequests",
	}
}

type GitLabClient interface {
	SearchIssues(query url.Values) ([]Issue, error)
	GetRepo() (Repository, error)
	SearchMergeRequests(query url.Values) ([]MergeRequest, error)
	SearchTags(query url.Values) ([]Tag, error)
	FindCommitRefsBySHA(sha string, query url.Values) ([]CommitRef, error)
	CreateTagRelease(tagName string, body Release) error
	UpdateTagRelease(tagName string, body Release) error
}

type Issue struct {
	IID    int    `json:"iid"`
	Title  string `json:"title"`
	WebUrl string `json:"web_url"`
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
