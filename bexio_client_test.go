package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestBexioClientCreateTimesheet(t *testing.T) {
	t.Parallel()

	request := bexioCreateTimesheetRequest{
		UserID:          42,
		AllowableBill:   true,
		ClientServiceID: 99,
		Text:            "Build acceptance test",
		ContactID:       intPtr(11),
		PrProjectID:     intPtr(12),
		Tracking: trackingRange{
			Type:  "range",
			Date:  "2026-02-08",
			Start: "09:00",
			End:   "10:30",
		},
	}

	created := bexioTimesheet{
		ID:              777,
		UserID:          request.UserID,
		AllowableBill:   request.AllowableBill,
		ClientServiceID: request.ClientServiceID,
		Text:            request.Text,
		ContactID:       request.ContactID,
		PrProjectID:     request.PrProjectID,
		Tracking:        request.Tracking,
	}

	responseJSON, err := json.Marshal(created)
	assert.NoError(t, err)

	received := fakeBexioCapturedRequest{}
	server := newRawResponseServer(t, &received, string(responseJSON), captureTimesheetBody)

	client := NewBexioClient(server.URL, "test-token", http.DefaultClient)

	result, err := client.CreateTimesheet(context.Background(), request)
	assert.NoError(t, err)
	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodPost,
		Path:          "/2.0/timesheet",
		Authorization: "Bearer test-token",
		Body:          request,
	}, received)
	assert.Equal(t, created, result)
}

func TestBexioClientDeleteTimesheet(t *testing.T) {
	t.Parallel()

	const responseBody = `{"success":true}`

	received := fakeBexioCapturedRequest{}
	server := newRawResponseServer(t, &received, responseBody, nil)

	client := NewBexioClient(server.URL, "test-token", http.DefaultClient)

	result, err := client.DeleteTimesheet(context.Background(), 777)
	assert.NoError(t, err)
	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodDelete,
		Path:          "/2.0/timesheet/777",
		Authorization: "Bearer test-token",
	}, received)
	assert.Equal(t, json.RawMessage(responseBody), result)
}

func TestBexioClientEditTimesheet(t *testing.T) {
	t.Parallel()

	request := bexioCreateTimesheetRequest{
		UserID:          42,
		AllowableBill:   true,
		ClientServiceID: 99,
		Text:            "Updated acceptance test",
		ContactID:       intPtr(11),
		PrProjectID:     intPtr(12),
		Tracking: trackingRange{
			Type:  "range",
			Date:  "2026-02-08",
			Start: "13:00",
			End:   "14:30",
		},
	}

	expected := bexioTimesheet{
		ID:              777,
		UserID:          request.UserID,
		AllowableBill:   request.AllowableBill,
		ClientServiceID: request.ClientServiceID,
		Text:            request.Text,
		ContactID:       request.ContactID,
		PrProjectID:     request.PrProjectID,
		Tracking:        request.Tracking,
	}

	responseJSON, err := json.Marshal(expected)
	assert.NoError(t, err)

	received := fakeBexioCapturedRequest{}
	server := newRawResponseServer(t, &received, string(responseJSON), captureTimesheetBody)

	client := NewBexioClient(server.URL, "test-token", http.DefaultClient)

	result, err := client.EditTimesheet(context.Background(), 777, request)
	assert.NoError(t, err)
	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodPost,
		Path:          "/2.0/timesheet/777",
		Authorization: "Bearer test-token",
		Body:          request,
	}, received)
	assert.Equal(t, expected, result)
}

func TestBexioClientListTimesheets(t *testing.T) {
	t.Parallel()

	expected := listTimesheetEntriesFixture()

	responseJSON, err := json.Marshal(expected)
	assert.NoError(t, err)

	received := fakeBexioCapturedRequest{}
	server := newRawResponseServer(t, &received, string(responseJSON), nil)

	client := NewBexioClient(server.URL, "test-token", http.DefaultClient)

	result, err := client.ListTimesheets(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodGet,
		Path:          "/2.0/timesheet",
		Query:         "limit=2000&order_by=date_desc",
		Authorization: "Bearer test-token",
	}, received)
	assert.Equal(t, expected, result)
}

func TestBexioClientSearchTimesheets(t *testing.T) {
	t.Parallel()

	searchFields := []bexioSearchField{
		{Field: "user_id", Value: "42", Criteria: "="},
	}

	expected := searchTimesheetEntriesFixture()

	responseJSON, err := json.Marshal(expected)
	assert.NoError(t, err)

	received := fakeBexioCapturedRequest{}
	server := newRawResponseServer(t, &received, string(responseJSON), captureSearchBody)

	client := NewBexioClient(server.URL, "test-token", http.DefaultClient)

	result, err := client.SearchTimesheets(context.Background(), searchFields)
	assert.NoError(t, err)
	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodPost,
		Path:          "/2.0/timesheet/search",
		Query:         "limit=2000&order_by=date_desc",
		Authorization: "Bearer test-token",
		SearchBody:    searchFields,
	}, received)
	assert.Equal(t, expected, result)
}

func TestBexioClientListContacts(t *testing.T) {
	t.Parallel()

	const responseBody = `[{"id":1,"name_1":"Acme Corp","name_2":""}]`
	expected := []bexioContact{{
		ID:    1,
		Name1: "Acme Corp",
		Name2: "",
	}}

	received := fakeBexioCapturedRequest{}
	server := newRawResponseServer(t, &received, responseBody, nil)

	client := NewBexioClient(server.URL, "test-token", http.DefaultClient)

	result, err := client.ListContacts(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodGet,
		Path:          "/2.0/contact",
		Authorization: "Bearer test-token",
	}, received)
	assert.Equal(t, expected, result)
}

func TestBexioClientSendsAcceptHeader(t *testing.T) {
	t.Parallel()

	var gotAccept string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAccept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`[]`))
		assert.NoError(t, err)
	}))
	t.Cleanup(server.Close)

	client := NewBexioClient(server.URL, "test-token", http.DefaultClient)

	_, err := client.ListContacts(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "application/json", gotAccept)
}

func TestBexioClientListProjects(t *testing.T) {
	t.Parallel()

	const responseBody = `[{"id":501,"name":"Project Alpha","contact_id":11}]`
	expected := []bexioProject{{
		ID:        501,
		Name:      "Project Alpha",
		ContactID: 11,
	}}

	received := fakeBexioCapturedRequest{}
	server := newRawResponseServer(t, &received, responseBody, nil)

	client := NewBexioClient(server.URL, "test-token", http.DefaultClient)

	result, err := client.ListProjects(context.Background())
	assert.NoError(t, err)
	var typedProjects []bexioProject = result
	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodGet,
		Path:          "/2.0/pr_project",
		Authorization: "Bearer test-token",
	}, received)
	assert.Equal(t, expected, typedProjects)
}

func TestBexioClientSearchProjects(t *testing.T) {
	t.Parallel()

	const responseBody = `[{"id":501,"name":"Project Alpha","contact_id":11}]`

	contactID := 11
	expected := []bexioProject{{
		ID:        501,
		Name:      "Project Alpha",
		ContactID: 11,
	}}
	expectedSearchBody := []bexioSearchField{{Field: "contact_id", Value: "11", Criteria: "="}}

	received := fakeBexioCapturedRequest{}
	server := newRawResponseServer(
		t,
		&received,
		responseBody,
		func(t *testing.T, r *http.Request, captured *fakeBexioCapturedRequest) {
			var body []bexioSearchField
			err := json.NewDecoder(r.Body).Decode(&body)
			assert.NoError(t, err)
			captured.SearchBody = body
		},
	)

	client := NewBexioClient(server.URL, "test-token", http.DefaultClient)

	result, err := client.SearchProjects(context.Background(), contactID)
	assert.NoError(t, err)
	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodPost,
		Path:          "/2.0/pr_project/search",
		Authorization: "Bearer test-token",
		SearchBody:    expectedSearchBody,
	}, received)
	assert.Equal(t, expected, result)
}

func TestBexioClientListClientServices(t *testing.T) {
	t.Parallel()

	const responseBody = `[{"id":77,"name":"Engineering"},{"id":78,"name":"Consulting"}]`
	expected := []bexioService{
		{ID: 77, Name: "Engineering"},
		{ID: 78, Name: "Consulting"},
	}

	received := fakeBexioCapturedRequest{}
	server := newRawResponseServer(t, &received, responseBody, nil)

	client := NewBexioClient(server.URL, "test-token", http.DefaultClient)

	result, err := client.ListClientServices(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodGet,
		Path:          "/2.0/client_service",
		Authorization: "Bearer test-token",
	}, received)
	assert.Equal(t, expected, result)
}

func TestBexioClientListPackages(t *testing.T) {
	t.Parallel()

	const responseBody = `[{"id":61,"name":"Backend Sprint"},{"id":62,"name":"QA Run"}]`
	expected := []bexioPackage{
		{ID: 61, Name: "Backend Sprint"},
		{ID: 62, Name: "QA Run"},
	}

	received := fakeBexioCapturedRequest{}
	server := newRawResponseServer(t, &received, responseBody, nil)

	client := NewBexioClient(server.URL, "test-token", http.DefaultClient)

	result, err := client.ListPackages(context.Background(), 5)
	assert.NoError(t, err)
	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodGet,
		Path:          "/3.0/projects/5/packages",
		Authorization: "Bearer test-token",
	}, received)
	assert.Equal(t, expected, result)
}

func TestBexioClientReturnsErrorOnNon2xxStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error":"invalid field"}`))
	}))
	t.Cleanup(server.Close)

	client := NewBexioClient(server.URL, "test-token", http.DefaultClient)

	_, err := client.CreateTimesheet(context.Background(), bexioCreateTimesheetRequest{
		UserID:          1,
		ClientServiceID: 1,
		Tracking:        trackingRange{Type: "range", Start: "09:00", End: "10:00"},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status 422")
	assert.Contains(t, err.Error(), "invalid field")
}

func captureTimesheetBody(t *testing.T, r *http.Request, captured *fakeBexioCapturedRequest) {
	t.Helper()

	var body bexioCreateTimesheetRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	assert.NoError(t, err)
	captured.Body = body
}

func captureSearchBody(t *testing.T, r *http.Request, captured *fakeBexioCapturedRequest) {
	t.Helper()

	var body []bexioSearchField
	err := json.NewDecoder(r.Body).Decode(&body)
	assert.NoError(t, err)
	captured.SearchBody = body
}

func newRawResponseServer(
	t *testing.T,
	received *fakeBexioCapturedRequest,
	responseBody string,
	readBody func(t *testing.T, r *http.Request, captured *fakeBexioCapturedRequest),
) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		captured := fakeBexioCapturedRequest{
			Method:        r.Method,
			Path:          r.URL.Path,
			Query:         r.URL.RawQuery,
			Authorization: r.Header.Get("Authorization"),
		}

		if readBody != nil {
			readBody(t, r, &captured)
		}

		*received = captured

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(responseBody))
		assert.NoError(t, err)
	}))

	t.Cleanup(server.Close)
	return server
}
