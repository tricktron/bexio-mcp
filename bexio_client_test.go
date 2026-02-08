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
	tests := []struct {
		name string
	}{
		{name: "posts timesheet with bearer token and returns created entry"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
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

			received := fakeBexioCapturedRequest{}
			server := newTimesheetServer(t, &received, created)

			client := NewBexioClient(server.URL, "test-token", http.DefaultClient)

			result, err := client.CreateTimesheet(context.Background(), request)
			assert.NoError(t, err)
			assert.Equal(t, http.MethodPost, received.Method)
			assert.Equal(t, "/2.0/timesheet", received.Path)
			assert.Equal(t, "Bearer test-token", received.Authorization)
			assert.Equal(t, request, received.Body)
			assert.Equal(t, created, result)
		})
	}
}

func newTimesheetServer(t *testing.T, received *fakeBexioCapturedRequest, response bexioTimesheet) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var body bexioCreateTimesheetRequest
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)

		*received = fakeBexioCapturedRequest{
			Method:        r.Method,
			Path:          r.URL.Path,
			Authorization: r.Header.Get("Authorization"),
			Body:          body,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(response)
		assert.NoError(t, err)
	}))

	t.Cleanup(server.Close)
	return server
}
