package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const (
	timesheetEndpoint       = "/2.0/timesheet"
	projectEndpoint         = "/2.0/pr_project"
	contactEndpoint         = "/2.0/contact"
	clientServiceEndpoint   = "/2.0/client_service"
	timesheetStatusEndpoint = "/2.0/timesheet_status"
	currentUserEndpoint     = "/3.0/users/me"
	searchEndpointSuffix    = "/search"
)

type BexioClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewBexioClient(baseURL, token string, httpClient *http.Client) BexioClient {
	return BexioClient{
		baseURL:    baseURL,
		token:      token,
		httpClient: httpClient,
	}
}

func (c BexioClient) CreateTimesheet(ctx context.Context, req bexioCreateTimesheetRequest) (bexioTimesheet, error) {
	return c.postTimesheet(ctx, timesheetEndpoint, req)
}

func (c BexioClient) EditTimesheet(
	ctx context.Context,
	id int,
	req bexioCreateTimesheetRequest,
) (bexioTimesheet, error) {
	return c.postTimesheet(ctx, fmt.Sprintf("%s/%d", timesheetEndpoint, id), req)
}

func (c BexioClient) postTimesheet(
	ctx context.Context,
	endpoint string,
	req bexioCreateTimesheetRequest,
) (bexioTimesheet, error) {
	httpReq, err := c.newJSONRequest(ctx, endpoint, req)
	if err != nil {
		return bexioTimesheet{}, err
	}

	var timesheet bexioTimesheet
	err = c.doAndDecode(httpReq, &timesheet)
	if err != nil {
		return bexioTimesheet{}, err
	}

	return timesheet, nil
}

func (c BexioClient) DeleteTimesheet(ctx context.Context, id int) (json.RawMessage, error) {
	httpReq, err := c.newRequest(ctx, http.MethodDelete, fmt.Sprintf("%s/%d", timesheetEndpoint, id), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	return c.readRawResponse(httpReq)
}

func (c BexioClient) ListTimesheets(ctx context.Context) ([]bexioTimesheet, error) {
	httpReq, err := c.newRequest(ctx, http.MethodGet, timesheetEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	q := httpReq.URL.Query()
	q.Set("order_by", "date_desc")
	q.Set("limit", "2000")
	httpReq.URL.RawQuery = q.Encode()

	var timesheets []bexioTimesheet
	err = c.doAndDecode(httpReq, &timesheets)
	if err != nil {
		return nil, err
	}

	return timesheets, nil
}

func (c BexioClient) SearchTimesheets(ctx context.Context, fields []bexioSearchField) ([]bexioTimesheet, error) {
	httpReq, err := c.newJSONRequest(ctx, timesheetEndpoint+searchEndpointSuffix, fields)
	if err != nil {
		return nil, err
	}

	q := httpReq.URL.Query()
	q.Set("order_by", "date_desc")
	q.Set("limit", "2000")
	httpReq.URL.RawQuery = q.Encode()

	var timesheets []bexioTimesheet
	err = c.doAndDecode(httpReq, &timesheets)
	if err != nil {
		return nil, err
	}

	return timesheets, nil
}

func (c BexioClient) ListContacts(ctx context.Context) ([]bexioContact, error) {
	var contacts []bexioContact
	err := c.getAndDecode(ctx, contactEndpoint, &contacts)
	if err != nil {
		return nil, err
	}

	return contacts, nil
}

func (c BexioClient) ListProjects(ctx context.Context) ([]bexioProject, error) {
	var projects []bexioProject
	err := c.getAndDecode(ctx, projectEndpoint, &projects)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (c BexioClient) ListClientServices(ctx context.Context) ([]bexioService, error) {
	var services []bexioService
	err := c.getAndDecode(ctx, clientServiceEndpoint, &services)
	if err != nil {
		return nil, err
	}

	return services, nil
}

func (c BexioClient) ListTimesheetStatuses(ctx context.Context) ([]timesheetStatus, error) {
	var statuses []timesheetStatus
	err := c.getAndDecode(ctx, timesheetStatusEndpoint, &statuses)
	if err != nil {
		return nil, err
	}

	return statuses, nil
}

func (c BexioClient) GetCurrentUser(ctx context.Context) (bexioUser, error) {
	var user bexioUser
	err := c.getAndDecode(ctx, currentUserEndpoint, &user)
	if err != nil {
		return bexioUser{}, err
	}

	return user, nil
}

func (c BexioClient) ListPackages(ctx context.Context, projectID int) ([]bexioPackage, error) {
	var packages []bexioPackage
	err := c.getAndDecode(ctx, fmt.Sprintf("/3.0/projects/%d/packages", projectID), &packages)
	if err != nil {
		return nil, err
	}

	return packages, nil
}

func (c BexioClient) SearchProjects(ctx context.Context, contactID int) ([]bexioProject, error) {
	fields := []bexioSearchField{{Field: "contact_id", Value: strconv.Itoa(contactID), Criteria: "="}}
	httpReq, err := c.newJSONRequest(ctx, projectEndpoint+searchEndpointSuffix, fields)
	if err != nil {
		return nil, err
	}

	var projects []bexioProject
	err = c.doAndDecode(httpReq, &projects)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (c BexioClient) readRawResponse(httpReq *http.Request) (json.RawMessage, error) {
	body, err := c.doRequest(httpReq)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	raw, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return json.RawMessage(raw), nil
}

func (c BexioClient) newJSONRequest(ctx context.Context, path string, payload any) (*http.Request, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := c.newRequest(ctx, http.MethodPost, path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	return httpReq, nil
}

func (c BexioClient) doAndDecode(httpReq *http.Request, target any) error {
	body, err := c.doRequest(httpReq)
	if err != nil {
		return err
	}
	defer body.Close()

	if err = json.NewDecoder(body).Decode(target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func (c BexioClient) getAndDecode(ctx context.Context, endpoint string, target any) error {
	httpReq, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	return c.doAndDecode(httpReq, target)
}

func (c BexioClient) doRequest(httpReq *http.Request) (io.ReadCloser, error) {
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		defer httpResp.Body.Close()

		body, _ := io.ReadAll(httpResp.Body)

		return nil, fmt.Errorf(
			"bexio API %s %s: status %d: %s",
			httpReq.Method, httpReq.URL.Path,
			httpResp.StatusCode, body,
		)
	}

	return httpResp.Body, nil
}

func (c BexioClient) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	httpReq, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.token)
	httpReq.Header.Set("Accept", "application/json")

	return httpReq, nil
}
