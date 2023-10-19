package app

import (
	"fmt"
	"gitLab-rls-note/pkg/errors"
	"strings"
	"time"
)

var LABEL_CONFIG = []LabelConfig{
	{Name: "breaking change", Title: "Notable changes"},
	{Name: "enhancement", Title: "Enhancements"},
	{Name: "feature", Title: "New features"},
	{Name: "bug", Title: "Fixed bugs"},
	{Name: "issues", Title: "Closed issues"},
	{Name: "mergeRequests", Title: "Merged requests"},
}

const (
	yyyy_mm_dd = "2006-01-02"
)

type ContentService interface {
	GenerateContent(mergeReqs []MergeRequest, issues []Issue, releaseDate string) (string, error)
}
type contentService struct {
	labelConfigs []LabelConfig
	labelBucket  map[string][]string
	timeZone     *time.Location
}

func NewContentService(timeZone string) (ContentService, error) {
	labelBucket := make(map[string][]string)
	for _, item := range LABEL_CONFIG {
		labelBucket[item.Name] = []string{}
	}

	tz, err := time.LoadLocation(timeZone)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &contentService{LABEL_CONFIG, labelBucket, tz}, nil
}

func (s *contentService) GenerateContent(mergeReqs []MergeRequest, issues []Issue, releaseDate string) (string, error) {
	var dMrs []DecoratedMergeRequest
	for _, mr := range mergeReqs {
		dMrs = append(dMrs, s.decorateMergeRequest(mr))
	}

	var dIssues []DecoratedIssue
	for _, issue := range issues {
		dIssues = append(dIssues, s.decorateIssue(issue))
	}

	s.populateLabelBucket(dMrs, dIssues)

	return s.generateReleaseNote(releaseDate)
}

func (s *contentService) generateReleaseNote(releaseDate string) (string, error) {
	parsedReleaseDate, err := time.Parse(time.RFC3339Nano, releaseDate)
	if err != nil {
		return "", errors.WithStack(err)
	}

	output := fmt.Sprintf("### Release note (%s)\n", parsedReleaseDate.In(s.timeZone).Format(yyyy_mm_dd))
	for _, label := range s.labelConfigs {
		if _, exists := s.labelBucket[label.Name]; exists {
			bucket := s.labelBucket[label.Name]
			isEmpty := len(bucket) > 0
			if isEmpty {
				output += fmt.Sprintf("#### %s\n", label.Title) + strings.Join(bucket, "\n") + "\n"
			}
		}
	}
	return output, nil
}

func (s *contentService) populateLabelBucket(mergeReqs []DecoratedMergeRequest, issues []DecoratedIssue) {
	for _, mr := range mergeReqs {
		added := false
		for _, label := range mr.Labels {
			if _, exists := s.labelBucket[label]; exists {
				s.labelBucket[label] = append(s.labelBucket[label], mr.Message)
				added = true
			}
		}

		if !added {
			s.labelBucket[mr.DefaultLabel] = append(s.labelBucket[mr.DefaultLabel], mr.Message)
		}
	}

	for _, issue := range issues {
		added := false
		for _, label := range issue.Labels {
			if _, exists := s.labelBucket[label]; exists {
				s.labelBucket[label] = append(s.labelBucket[label], issue.Message)
				added = true
			}
		}

		if !added {
			s.labelBucket[issue.DefaultLabel] = append(s.labelBucket[issue.DefaultLabel], issue.Message)
		}
	}
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

type LabelConfig struct {
	Name  string
	Title string
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
