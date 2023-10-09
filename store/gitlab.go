package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"gitLab-rls-note/app"
	"gitLab-rls-note/pkg/errors"
)

type gitlabClient struct {
	personalToken string
	apiEndpoint   string
	projectID     string
}

func NewGitlabClient(personalToken, apiEndpoint, projectID string) app.GitLabClient {
	return &gitlabClient{
		personalToken: personalToken,
		apiEndpoint:   apiEndpoint,
		projectID:     projectID,
	}
}

func (g *gitlabClient) RetrieveIssues(query url.Values) ([]app.Issue, error) {
	path := fmt.Sprintf("/projects/%s/issues", g.projectID)
	body, err := g.makeRequest(requestIn{method: "GET", path: path, query: query})
	if err != nil {
		return nil, err
	}

	var issues []app.Issue
	if err := json.Unmarshal(body, &issues); err != nil {
		return nil, errors.WithStack(err)
	}

	return issues, nil
}

func (g *gitlabClient) RetrieveRepo() (app.Repository, error) {
	path := fmt.Sprintf("/projects/%s", g.projectID)
	body, err := g.makeRequest(requestIn{method: "GET", path: path})
	if err != nil {
		return app.Repository{}, err
	}

	var repo app.Repository
	if err := json.Unmarshal(body, &repo); err != nil {
		return app.Repository{}, errors.WithStack(err)
	}

	return repo, nil
}

func (g *gitlabClient) RetrieveMergeRequests(query url.Values) ([]app.MergeRequest, error) {
	path := fmt.Sprintf("/projects/%s/merge_requests", g.projectID)
	body, err := g.makeRequest(requestIn{method: "GET", path: path, query: query})
	if err != nil {
		return nil, err
	}

	var mergeRequests []app.MergeRequest
	if err := json.Unmarshal(body, &mergeRequests); err != nil {
		return nil, errors.WithStack(err)
	}

	return mergeRequests, nil
}

func (g *gitlabClient) RetrieveTags(query url.Values) ([]app.Tag, error) {
	path := fmt.Sprintf("/projects/%s/repository/tags", g.projectID)
	body, err := g.makeRequest(requestIn{method: "GET", path: path, query: query})
	if err != nil {
		return nil, err
	}

	var tags []app.Tag
	if err := json.Unmarshal(body, &tags); err != nil {
		return nil, errors.WithStack(err)
	}

	return tags, nil
}

func (g *gitlabClient) RetrieveCommitRefsBySHA(sha string, query url.Values) ([]app.CommitRef, error) {
	path := fmt.Sprintf("/projects/%s/repository/commits/%s/refs", g.projectID, sha)
	body, err := g.makeRequest(requestIn{method: "GET", path: path, query: query})
	if err != nil {
		return nil, err
	}

	var commitRefs []app.CommitRef
	if err := json.Unmarshal(body, &commitRefs); err != nil {
		return nil, errors.WithStack(err)
	}

	return commitRefs, nil
}

func (g *gitlabClient) CreateTagRelease(tagName string, body app.Release) error {
	path := fmt.Sprintf("/projects/%s/releases", g.projectID)
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return errors.WithStack(err)
	}

	query := url.Values{"tag_name": {tagName}}
	_, err = g.makeRequest(requestIn{method: "POST", path: path, query: query, body: bodyJSON})
	if err != nil {
		return err
	}

	return nil
}

func (g *gitlabClient) UpdateTagRelease(tagName string, body app.Release) error {
	path := fmt.Sprintf("/projects/%s/releases/%s", g.projectID, tagName)
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = g.makeRequest(requestIn{method: "PUT", path: path, body: bodyJSON})
	if err != nil {
		return err
	}

	return nil
}

func (g *gitlabClient) makeRequest(reqIn requestIn) ([]byte, error) {
	client := &http.Client{}
	fullURL := fmt.Sprintf("%s%s?%s", g.apiEndpoint, reqIn.path, reqIn.query.Encode())

	req, err := http.NewRequest(reqIn.method, fullURL, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req.Header.Add("Private-Token", g.personalToken)

	if reqIn.body != nil {
		req.Body = io.NopCloser(bytes.NewReader(reqIn.body))
		req.Header.Add("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("HTTP request failed with status code %d", resp.StatusCode))
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return responseBody, nil
}

type requestIn struct {
	method string
	path   string
	query  url.Values
	body   []byte
}
