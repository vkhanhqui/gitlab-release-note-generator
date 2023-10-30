package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"gitLab-rls-note/app"
	"gitLab-rls-note/pkg/errors"

	"strconv"
)

const (
	GitlabTimeFormat = time.RFC3339Nano
)

type gitlabClient struct {
	personalToken string
	apiEndpoint   string
	projectID     string
	cookie        string
}

func NewGitlabClient(personalToken, apiEndpoint, projectID, cookie string) app.GitLabClient {
	return &gitlabClient{
		personalToken: personalToken,
		apiEndpoint:   apiEndpoint,
		projectID:     projectID,
		cookie:        cookie,
	}
}

func (g *gitlabClient) RetrieveIssues(prs app.ListIssueParams, pg *app.Pagination) ([]app.Issue, error) {
	path := fmt.Sprintf("/projects/%s/issues", g.projectID)
	query := url.Values{
		"updated_before": {prs.UpdatedBefore.Format(GitlabTimeFormat)},
		"updated_after":  {prs.UpdatedAfter.Format(GitlabTimeFormat)},
		"state":          {prs.State},
		"page":           {strconv.Itoa(pg.Page)},
		"per_page":       {strconv.Itoa(pg.PerPage)},
	}
	header, body, err := g.makeRequest(requestIn{method: http.MethodGet, path: path, query: query})
	if err != nil {
		return nil, err
	}

	var issues []app.Issue
	if err := json.Unmarshal(body, &issues); err != nil {
		return nil, errors.WithStack(err)
	}

	pg.Page = g.getNextPage(header)
	return issues, nil
}

func (g *gitlabClient) RetrieveRepo() (app.Repo, error) {
	path := fmt.Sprintf("/projects/%s", g.projectID)
	_, body, err := g.makeRequest(requestIn{method: http.MethodGet, path: path})
	if err != nil {
		return app.Repo{}, err
	}

	var repo app.Repo
	if err := json.Unmarshal(body, &repo); err != nil {
		return app.Repo{}, errors.WithStack(err)
	}

	return repo, nil
}

func (g *gitlabClient) RetrieveMergeRequests(prs app.ListMReqParams, pg *app.Pagination) ([]app.MergeRequest, error) {
	path := fmt.Sprintf("/projects/%s/merge_requests", g.projectID)

	query := url.Values{
		"target_branch":  {prs.TargetBranch},
		"scope":          {"all"},
		"updated_before": {prs.UpdatedBefore.Format(GitlabTimeFormat)},
		"updated_after":  {prs.UpdatedAfter.Format(GitlabTimeFormat)},
		"state":          {prs.State},
		"page":           {strconv.Itoa(pg.Page)},
		"per_page":       {strconv.Itoa(pg.PerPage)},
	}
	header, body, err := g.makeRequest(requestIn{method: http.MethodGet, path: path, query: query})
	if err != nil {
		return nil, err
	}

	var mergeRequests []app.MergeRequest
	if err := json.Unmarshal(body, &mergeRequests); err != nil {
		return nil, errors.WithStack(err)
	}

	pg.Page = g.getNextPage(header)
	return mergeRequests, nil
}

func (g *gitlabClient) RetrieveTags(pg *app.Pagination) ([]app.Tag, error) {
	path := fmt.Sprintf("/projects/%s/repository/tags", g.projectID)
	query := url.Values{
		"page":     {strconv.Itoa(pg.Page)},
		"per_page": {strconv.Itoa(pg.PerPage)},
	}
	header, body, err := g.makeRequest(requestIn{method: http.MethodGet, path: path, query: query})
	if err != nil {
		return nil, err
	}

	var tags []app.Tag
	if err := json.Unmarshal(body, &tags); err != nil {
		return nil, errors.WithStack(err)
	}

	pg.Page = g.getNextPage(header)
	return tags, nil
}

func (g *gitlabClient) RetrieveCommitRefsBySHA(sha string, query url.Values) ([]app.CommitRef, error) {
	path := fmt.Sprintf("/projects/%s/repository/commits/%s/refs", g.projectID, sha)
	_, body, err := g.makeRequest(requestIn{method: http.MethodGet, path: path, query: query})
	if err != nil {
		return nil, err
	}

	var commitRefs []app.CommitRef
	if err := json.Unmarshal(body, &commitRefs); err != nil {
		return nil, errors.WithStack(err)
	}

	return commitRefs, nil
}

func (g *gitlabClient) CreateTagRelease(body app.Release) error {
	path := fmt.Sprintf("/projects/%s/releases", g.projectID)
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return errors.WithStack(err)
	}

	_, _, err = g.makeRequest(requestIn{method: http.MethodPost, path: path, body: bodyJSON})
	if err != nil {
		return err
	}

	return nil
}

func (g *gitlabClient) UpdateTagRelease(body app.Release) error {
	path := fmt.Sprintf("/projects/%s/releases/%s", g.projectID, body.Name)
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return errors.WithStack(err)
	}

	_, _, err = g.makeRequest(requestIn{method: http.MethodPut, path: path, body: bodyJSON})
	if err != nil {
		return err
	}

	return nil
}

func (g *gitlabClient) makeRequest(reqIn requestIn) (http.Header, []byte, error) {
	client := &http.Client{}
	fullURL := fmt.Sprintf("%s%s?%s", g.apiEndpoint, reqIn.path, reqIn.query.Encode())

	req, err := http.NewRequest(reqIn.method, fullURL, nil)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	req.Header.Add("Private-Token", g.personalToken)

	if g.cookie != "" {
		req.Header.Add("cookie", g.cookie)
	}

	if reqIn.body != nil {
		req.Body = io.NopCloser(bytes.NewReader(reqIn.body))
		req.Header.Add("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	return resp.Header, responseBody, nil
}

type requestIn struct {
	method string
	path   string
	query  url.Values
	body   []byte
}

func (g *gitlabClient) getNextPage(header http.Header) int {
	nextPageStr := header.Get("X-Next-Page")
	nextPage, err := strconv.Atoi(nextPageStr)
	if err != nil {
		nextPage = app.GitLabDefaultPage
	}
	return nextPage
}
