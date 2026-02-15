package main

import (
	"bytes"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestStartupValidationAcceptance(t *testing.T) {
	// Slice: Startup validation and fail-fast
	testCases := []struct {
		name          string
		token         string
		baseURL       string
		expectExit    int
		errorContains string
	}{
		{
			name:          "Given BEXIO_API_TOKEN is empty, when the MCP server starts, then it exits with a non-zero code and a message mentioning BEXIO_API_TOKEN",
			token:         "",
			baseURL:       "https://api.bexio.com",
			expectExit:    1,
			errorContains: "BEXIO_API_TOKEN",
		},
		{
			name:          "Given BEXIO_API_BASE_URL is empty, when the MCP server starts, then it exits with a non-zero code and a message mentioning BEXIO_API_BASE_URL",
			token:         "test-token",
			baseURL:       "",
			expectExit:    1,
			errorContains: "BEXIO_API_BASE_URL",
		},
		{
			name:          "Given both env vars are set and valid, when the MCP server starts, then it proceeds normally",
			token:         "test-token",
			baseURL:       "https://api.bexio.com",
			expectExit:    0,
			errorContains: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getenv := func(key string) string {
				switch key {
				case "BEXIO_API_TOKEN":
					return tc.token
				case "BEXIO_API_BASE_URL":
					return tc.baseURL
				default:
					return ""
				}
			}

			var stderr bytes.Buffer

			exitCode := run(getenv, &stderr)

			assert.Equal(t, tc.expectExit, exitCode)
			if tc.errorContains == "" {
				assert.Equal(t, "", stderr.String())
				return
			}

			assert.Contains(t, stderr.String(), tc.errorContains)
		})
	}
}
