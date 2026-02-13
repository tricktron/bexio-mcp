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
	httpReq, err := c.newJSONRequest(ctx, http.MethodPost, endpoint, req)
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

	var timesheets []bexioTimesheet
	err = c.doAndDecode(httpReq, &timesheets)
	if err != nil {
		return nil, err
	}

	return timesheets, nil
}

func (c BexioClient) SearchTimesheets(ctx context.Context, fields []bexioSearchField) ([]bexioTimesheet, error) {
	httpReq, err := c.newJSONRequest(ctx, http.MethodPost, timesheetEndpoint+searchEndpointSuffix, fields)
	if err != nil {
		return nil, err
	}

	var timesheets []bexioTimesheet
	err = c.doAndDecode(httpReq, &timesheets)
	if err != nil {
		return nil, err
	}

	return timesheets, nil
}

func (c BexioClient) ListContacts(ctx context.Context) (json.RawMessage, error) {
	return c.getRaw(ctx, contactEndpoint)
}

func (c BexioClient) ListClientServices(ctx context.Context) (json.RawMessage, error) {
	return c.getRaw(ctx, clientServiceEndpoint)
}

func (c BexioClient) ListTimesheetStatuses(ctx context.Context) (json.RawMessage, error) {
	return c.getRaw(ctx, timesheetStatusEndpoint)
}

func (c BexioClient) GetCurrentUser(ctx context.Context) (json.RawMessage, error) {
	return c.getRaw(ctx, currentUserEndpoint)
}

func (c BexioClient) ListPackages(ctx context.Context, projectID int) (json.RawMessage, error) {
	return c.getRaw(ctx, fmt.Sprintf("/3.0/projects/%d/packages", projectID))
}

func (c BexioClient) SearchProjects(ctx context.Context, contactID *int) (json.RawMessage, error) {
	fields := make([]bexioSearchField, 0, 1)
	if contactID != nil {
		fields = append(fields, bexioSearchField{Field: "contact_id", Value: strconv.Itoa(*contactID), Criteria: "="})
	}

	return c.postRaw(ctx, projectEndpoint+searchEndpointSuffix, fields)
}

func (c BexioClient) getRaw(ctx context.Context, endpoint string) (json.RawMessage, error) {
	httpReq, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	return c.readRawResponse(httpReq)
}

func (c BexioClient) postRaw(ctx context.Context, endpoint string, payload any) (json.RawMessage, error) {
	httpReq, err := c.newJSONRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return nil, err
	}

	return c.readRawResponse(httpReq)
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

func (c BexioClient) newJSONRequest(ctx context.Context, method, path string, payload any) (*http.Request, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := c.newRequest(ctx, method, path, bytes.NewReader(body))
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

	err = decodeJSON(body, target)
	if err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
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

	return httpReq, nil
}

func decodeJSON(src io.Reader, target any) error {
	return json.NewDecoder(src).Decode(target)
}
