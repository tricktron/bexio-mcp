package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

func TestCreateTimesheetAcceptance(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)

	fakeAPI := startFakeBexioAPI(t)
	mcp := startMCPServer(ctx, t, map[string]string{
		"BEXIO_API_TOKEN":    "test-token",
		"BEXIO_API_BASE_URL": fakeAPI.URL,
	})

	mcp.request(ctx, jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      intPtr(1),
		Method:  "initialize",
		Params: initializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]any{},
			ClientInfo: clientInfo{
				Name:    "acceptance-test",
				Version: "0.1.0",
			},
		},
	})
	mcp.notify(jsonRPCRequest{JSONRPC: "2.0", Method: "notifications/initialized"})

	response := mcp.request(ctx, jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      intPtr(2),
		Method:  "tools/call",
		Params: toolCallParams{
			Name: "create_timesheet",
			Arguments: bexioCreateTimesheetRequest{
				UserID:          42,
				AllowableBill:   true,
				ClientServiceID: 99,
				ContactID:       intPtr(11),
				PrProjectID:     intPtr(12),
				Text:            "Build acceptance test",
				Tracking: trackingRange{
					Type:  "range",
					Date:  "2026-02-08",
					Start: "09:00",
					End:   "10:30",
				},
			},
		},
	})

	received := fakeAPI.Received()
	assert.Equal(t, http.MethodPost, received.Method)
	assert.Equal(t, "/2.0/timesheet", received.Path)
	assert.Equal(t, "Bearer test-token", received.Authorization)
	assert.Equal(t, 99, received.Body.ClientServiceID)
	assert.Equal(t, "Build acceptance test", received.Body.Text)
	assert.Equal(t, "range", received.Body.Tracking.Type)
	assert.Equal(t, "09:00", received.Body.Tracking.Start)
	assert.Equal(t, "10:30", received.Body.Tracking.End)

	var toolResult toolCallResult
	err := json.Unmarshal(response.Result, &toolResult)
	assert.NoError(t, err)
	assert.Equal(t, 777, toolResult.StructuredContent.ID)
}

type clientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type initializeParams struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ClientInfo      clientInfo     `json:"clientInfo"`
}

type toolCallParams struct {
	Name      string                      `json:"name"`
	Arguments bexioCreateTimesheetRequest `json:"arguments"`
}

type jsonRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      *int   `json:"id,omitempty"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type toolCallResult struct {
	StructuredContent bexioTimesheet `json:"structuredContent"`
}

type fakeBexioCapturedRequest struct {
	Method        string
	Path          string
	Authorization string
	Body          bexioCreateTimesheetRequest
}

type fakeBexioAPI struct {
	URL    string
	server *httptest.Server

	mu       sync.Mutex
	captured fakeBexioCapturedRequest
}

func startFakeBexioAPI(t *testing.T) *fakeBexioAPI {
	t.Helper()

	api := &fakeBexioAPI{}
	api.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var reqBody bexioCreateTimesheetRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatal(err)
		}

		api.mu.Lock()
		api.captured = fakeBexioCapturedRequest{
			Method:        r.Method,
			Path:          r.URL.Path,
			Authorization: r.Header.Get("Authorization"),
			Body:          reqBody,
		}
		api.mu.Unlock()

		created := bexioTimesheet{
			ID:              777,
			UserID:          reqBody.UserID,
			AllowableBill:   reqBody.AllowableBill,
			ClientServiceID: reqBody.ClientServiceID,
			Text:            reqBody.Text,
			ContactID:       reqBody.ContactID,
			PrProjectID:     reqBody.PrProjectID,
			Tracking:        reqBody.Tracking,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(created); err != nil {
			t.Fatal(err)
		}
	}))
	api.URL = api.server.URL
	t.Cleanup(api.server.Close)

	return api
}

func (f *fakeBexioAPI) Received() fakeBexioCapturedRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.captured
}

type mcpServer struct {
	t      *testing.T
	stdin  io.WriteCloser
	stdout *bufio.Reader
	stderr *bytes.Buffer
	cmd    *exec.Cmd
}

func startMCPServer(ctx context.Context, t *testing.T, env map[string]string) *mcpServer {
	t.Helper()

	cmd := exec.CommandContext(ctx, "go", "run", ".")
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	in, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr

	err = cmd.Start()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = in.Close()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	})

	return &mcpServer{t: t, stdin: in, stdout: bufio.NewReader(out), stderr: stderr, cmd: cmd}
}

func (m *mcpServer) notify(msg jsonRPCRequest) {
	m.t.Helper()
	writeRPC(m.t, m.stdin, msg)
}

func (m *mcpServer) request(ctx context.Context, msg jsonRPCRequest) jsonRPCResponse {
	m.t.Helper()
	writeRPC(m.t, m.stdin, msg)

	for {
		resp := readRPC(ctx, m.t, m.stdout, m.stderr)
		if msg.ID != nil && resp.ID != nil && *msg.ID == *resp.ID {
			if resp.Error != nil {
				m.t.Fatalf("json-rpc error: %d %s", resp.Error.Code, resp.Error.Message)
			}
			return resp
		}
	}
}

func writeRPC(t *testing.T, w io.Writer, msg jsonRPCRequest) {
	t.Helper()

	payload, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	header := "Content-Length: " + strconv.Itoa(len(payload)) + "\r\n\r\n"
	_, err = io.WriteString(w, header)
	if err != nil {
		t.Fatal(err)
	}
	_, err = w.Write(payload)
	if err != nil {
		t.Fatal(err)
	}
}

func readRPC(ctx context.Context, t *testing.T, r *bufio.Reader, stderr *bytes.Buffer) jsonRPCResponse {
	t.Helper()

	respCh := make(chan jsonRPCResponse, 1)
	errCh := make(chan error, 1)
	go func() {
		resp, err := readRPCOnce(r)
		if err != nil {
			errCh <- err
			return
		}
		respCh <- resp
	}()

	select {
	case <-ctx.Done():
		t.Fatalf("timed out waiting for MCP response: %v, stderr: %s", ctx.Err(), strings.TrimSpace(stderr.String()))
	case err := <-errCh:
		t.Fatalf("read MCP response: %v, stderr: %s", err, strings.TrimSpace(stderr.String()))
	case resp := <-respCh:
		return resp
	}
	return jsonRPCResponse{}
}

func readRPCOnce(r *bufio.Reader) (jsonRPCResponse, error) {
	length := 0
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return jsonRPCResponse{}, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 && strings.EqualFold(strings.TrimSpace(parts[0]), "Content-Length") {
			length, err = strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return jsonRPCResponse{}, err
			}
		}
	}

	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return jsonRPCResponse{}, err
	}

	var resp jsonRPCResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		return jsonRPCResponse{}, err
	}

	return resp, nil
}

func intPtr(v int) *int {
	return &v
}
