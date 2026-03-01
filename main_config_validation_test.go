package main

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestValidateConfigReturnsErrorForMissingRequiredEnv(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		token        string
		baseURL      string
		wantContains string
	}{
		{
			name:         "empty token",
			token:        "",
			baseURL:      "https://api.bexio.com",
			wantContains: "BEXIO_API_TOKEN",
		},
		{
			name:         "empty base URL",
			token:        "valid-token",
			baseURL:      "",
			wantContains: "BEXIO_API_BASE_URL",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validateConfig(tc.token, tc.baseURL)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantContains)
		})
	}
}
