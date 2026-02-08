package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
	body, err := json.Marshal(req)
	if err != nil {
		return bexioTimesheet{}, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/2.0/timesheet", bytes.NewReader(body))
	if err != nil {
		return bexioTimesheet{}, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.token)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return bexioTimesheet{}, fmt.Errorf("send request: %w", err)
	}
	defer httpResp.Body.Close()

	var created bexioTimesheet
	if err := json.NewDecoder(httpResp.Body).Decode(&created); err != nil {
		return bexioTimesheet{}, fmt.Errorf("decode response: %w", err)
	}

	return created, nil
}
