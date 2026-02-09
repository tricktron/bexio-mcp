package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const timesheetEndpoint = "/2.0/timesheet"

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
	body, err := json.Marshal(req)
	if err != nil {
		return bexioTimesheet{}, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := c.newRequest(ctx, http.MethodPost, timesheetEndpoint, bytes.NewReader(body))
	if err != nil {
		return bexioTimesheet{}, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return bexioTimesheet{}, fmt.Errorf("send request: %w", err)
	}
	defer httpResp.Body.Close()

	var created bexioTimesheet
	decodeErr := decodeJSON(httpResp.Body, &created)
	if decodeErr != nil {
		return bexioTimesheet{}, fmt.Errorf("decode response: %w", decodeErr)
	}

	return created, nil
}

func (c BexioClient) ListTimesheets(ctx context.Context) ([]bexioTimesheet, error) {
	httpReq, err := c.newRequest(ctx, http.MethodGet, timesheetEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer httpResp.Body.Close()

	var timesheets []bexioTimesheet
	decodeErr := decodeJSON(httpResp.Body, &timesheets)
	if decodeErr != nil {
		return nil, fmt.Errorf("decode response: %w", decodeErr)
	}

	return timesheets, nil
}

func (c BexioClient) SearchTimesheets(ctx context.Context, fields []bexioSearchField) ([]bexioTimesheet, error) {
	body, err := json.Marshal(fields)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := c.newRequest(ctx, http.MethodPost, timesheetEndpoint+"/search", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer httpResp.Body.Close()

	var timesheets []bexioTimesheet
	decodeErr := decodeJSON(httpResp.Body, &timesheets)
	if decodeErr != nil {
		return nil, fmt.Errorf("decode response: %w", decodeErr)
	}

	return timesheets, nil
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
