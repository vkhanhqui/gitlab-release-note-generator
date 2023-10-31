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

	GitLabDefaultPage    = 1
	GitLabDefaultPerPage = 20

	lookingSecondTagLimit = 100
)

type GitLabService interface {
	RetrieveTwoLatestTags() ([]Tag, error)
	RetrieveChangelogsByStartAndEndDate(startDate, endDate time.Time) ([]MergeRequest, []Issue, error)
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
	IncludeCommits     bool
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

func (s *gitLabService) RetrieveChangelogsByStartAndEndDate(startDate, endDate time.Time) ([]MergeRequest, []Issue, error) {
	mrs, err := s.retrieveMergeRequests(ListMReqParams{
		TargetBranch:  s.config.TargetBranch,
		UpdatedBefore: endDate,
		UpdatedAfter:  startDate,
		State:         mergeRequestState,
	})
	if err != nil {
		return nil, nil, err
	}

	var filteredMRs []MergeRequest
	for _, mr := range mrs {
		if mr.MergedAt.After(startDate) && mr.MergedAt.Before(endDate) {
			filteredMRs = append(filteredMRs, mr)
		}
	}

	if s.config.IncludeCommits {
		for i, mr := range filteredMRs {
			filteredMRs[i].Commits, err = s.retrieveMergeRequestCommits(mr.IID)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	issues, err := s.retrieveIssues(ListIssueParams{
		UpdatedBefore: endDate,
		UpdatedAfter:  startDate,
		State:         issueState,
	})
	if err != nil {
		return nil, nil, err
	}

	var filteredISs []Issue
	for _, iss := range issues {
		if iss.ClosedAt.After(startDate) && iss.ClosedAt.Before(endDate) {
			filteredISs = append(filteredISs, iss)
		}
	}

	return filteredMRs, filteredISs, nil
}

func (s *gitLabService) RetrieveTwoLatestTags() ([]Tag, error) {
	var pg Pagination
	pg.SetDefaults()
	tags, err := s.client.RetrieveTags(&pg)
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

	var secondTag Tag
	lookingLimit := lookingSecondTagLimit
	for secondTag.Name == "" && lookingLimit > 0 {
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
				secondTag = tag
				break
			}
		}

		if secondTag.Name == "" && pg.Page != GitLabDefaultPage {
			tags, err = s.client.RetrieveTags(&pg)
			if err != nil {
				return nil, err
			}
		}

		lookingLimit -= 1
	}

	if lookingLimit < 1 {
		return nil, errors.New("Cannot find latest and second latest tag. Abort the program!")
	}

	if s.config.IssueClosedSeconds > 0 {
		addedTime := time.Duration(s.config.IssueClosedSeconds) * time.Second
		latest.Commit.CommittedDate = latest.Commit.CommittedDate.Add(addedTime)
		secondTag.Commit.CommittedDate = secondTag.Commit.CommittedDate.Add(addedTime)
	}

	return []Tag{latest, secondTag}, nil
}

func (s *gitLabService) retrieveMergeRequests(prs ListMReqParams) ([]MergeRequest, error) {
	var pg Pagination
	pg.SetDefaults()
	var resp []MergeRequest
	mrs, err := s.client.RetrieveMergeRequests(prs, &pg)
	if err != nil {
		return nil, err
	}
	resp = append(resp, mrs...)

	for pg.Page != GitLabDefaultPage {
		mrs, err := s.client.RetrieveMergeRequests(prs, &pg)
		if err != nil {
			return nil, err
		}
		resp = append(resp, mrs...)
	}
	return resp, err
}

func (s *gitLabService) retrieveMergeRequestCommits(merge_request_iid int) ([]MRCommit, error) {
	var pg Pagination
	pg.SetDefaults()
	var resp []MRCommit
	commits, err := s.client.RetrieveMergeRequestCommits(merge_request_iid, &pg)
	resp = append(resp, commits...)

	for pg.Page != GitLabDefaultPage {
		commits, err := s.client.RetrieveMergeRequestCommits(merge_request_iid, &pg)
		if err != nil {
			return nil, err
		}
		resp = append(resp, commits...)
	}
	return resp, err
}

func (s *gitLabService) retrieveIssues(prs ListIssueParams) ([]Issue, error) {
	var pg Pagination
	pg.SetDefaults()
	var resp []Issue
	issues, err := s.client.RetrieveIssues(prs, &pg)
	if err != nil {
		return nil, err
	}
	resp = append(resp, issues...)

	for pg.Page != GitLabDefaultPage {
		issues, err := s.client.RetrieveIssues(prs, &pg)
		if err != nil {
			return nil, err
		}
		resp = append(resp, issues...)
	}
	return resp, err
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
	RetrieveIssues(prs ListIssueParams, pg *Pagination) ([]Issue, error)
	RetrieveRepo() (Repo, error)
	RetrieveMergeRequests(prs ListMReqParams, pg *Pagination) ([]MergeRequest, error)
	RetrieveMergeRequestCommits(merge_request_iid int, pg *Pagination) ([]MRCommit, error)
	RetrieveTags(pg *Pagination) ([]Tag, error)
	RetrieveCommitRefsBySHA(sha string, query url.Values) ([]CommitRef, error)
	CreateTagRelease(body Release) error
	UpdateTagRelease(body Release) error
}

type ListIssueParams struct {
	UpdatedBefore time.Time
	UpdatedAfter  time.Time
	State         string
}

type Issue struct {
	IID      int       `json:"iid"`
	Title    string    `json:"title"`
	WebURL   string    `json:"web_url"`
	Labels   []string  `json:"labels"`
	ClosedAt time.Time `json:"closed_at"`
}

type Repo struct {
	CreatedAt time.Time `json:"created_at"`
}

type ListMReqParams struct {
	TargetBranch  string
	UpdatedBefore time.Time
	UpdatedAfter  time.Time
	State         string
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
	MergedAt time.Time `json:"merged_at"`
	Commits  []MRCommit
}

type Tag struct {
	Name    string  `json:"name"`
	Commit  Commit  `json:"commit"`
	Release Release `json:"release"`
}

type Commit struct {
	ID            string    `json:"id"`
	CommittedDate time.Time `json:"committed_date"`
}

type CommitRef struct {
	Name string `json:"name"`
}

type Release struct {
	Name        string `json:"tag_name"`
	Description string `json:"description"`
}

type Pagination struct {
	Page    int
	PerPage int
}

func (p *Pagination) SetDefaults() {
	if p.Page < 1 {
		p.Page = GitLabDefaultPage
	}

	if p.PerPage < 20 {
		p.PerPage = GitLabDefaultPerPage
	}
}

type MRCommit struct {
	ID             string    `json:"id"`
	ShortID        string    `json:"short_id"`
	CreatedAt      time.Time `json:"created_at"`
	Title          string    `json:"title"`
	Message        string    `json:"message"`
	AuthorName     string    `json:"author_name"`
	AuthorEmail    string    `json:"author_email"`
	AuthoredDate   time.Time `json:"authored_date"`
	CommitterName  string    `json:"committer_name"`
	CommitterEmail string    `json:"committer_email"`
	CommittedDate  time.Time `json:"committed_date"`
	WebURL         string    `json:"web_url"`
}
