package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
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

			exitCode := run(getenv, &stderr, nil)

			assert.Equal(t, tc.expectExit, exitCode)
			if tc.errorContains == "" {
				assert.Equal(t, "", stderr.String())
				return
			}

			assert.Contains(t, stderr.String(), tc.errorContains)
		})
	}
}

func TestRunStartsServerAcceptance(t *testing.T) {
	// Slice: Wire run() to start the MCP server
	// Given valid env vars and an in-memory transport, when run() starts, then MCP initialize returns server info containing "bexio-mcp"
	fakeAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(fakeAPI.Close)

	getenv := func(key string) string {
		switch key {
		case "BEXIO_API_TOKEN":
			return "test-token"
		case "BEXIO_API_BASE_URL":
			return fakeAPI.URL
		default:
			return ""
		}
	}

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	var stderr bytes.Buffer
	exitCodeCh := make(chan int, 1)
	go func() {
		exitCodeCh <- run(getenv, &stderr, serverTransport)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	testClient := mcp.NewClient(&mcp.Implementation{Name: "acceptance-test", Version: "0.1.0"}, nil)
	clientSession, err := testClient.Connect(ctx, clientTransport, nil)
	assert.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, clientSession.Close())
	})

	initializeResult := clientSession.InitializeResult()
	assert.Equal(t, "bexio-mcp", initializeResult.ServerInfo.Name)

	select {
	case exitCode := <-exitCodeCh:
		assert.Equal(t, 0, exitCode)
		assert.Equal(t, "", stderr.String())
	default:
	}
}
