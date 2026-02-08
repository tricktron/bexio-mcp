package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type config struct {
	apiToken   string
	apiBaseURL string
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      *int      `json:"id,omitempty"`
	Result  any       `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcToolCallParams struct {
	Name      string                      `json:"name"`
	Arguments bexioCreateTimesheetRequest `json:"arguments"`
}

type rpcToolCallResult struct {
	StructuredContent bexioTimesheet `json:"structuredContent"`
}

type timesheetCreator interface {
	CreateTimesheet(ctx context.Context, req bexioCreateTimesheetRequest) (bexioTimesheet, error)
}

func main() {
	cfg := config{
		apiToken:   os.Getenv("BEXIO_API_TOKEN"),
		apiBaseURL: os.Getenv("BEXIO_API_BASE_URL"),
	}

	if err := run(os.Stdin, os.Stdout, cfg); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

func run(stdin io.Reader, stdout io.Writer, cfg config) error {
	client := NewBexioClient(cfg.apiBaseURL, cfg.apiToken, http.DefaultClient)

	reader := bufio.NewReader(stdin)
	for {
		payload, err := readPayload(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("read payload: %w", err)
		}

		var req rpcRequest
		err = json.Unmarshal(payload, &req)
		if err != nil {
			return fmt.Errorf("decode request: %w", err)
		}

		resp, shouldRespond := handleRequest(context.Background(), req, client)
		if !shouldRespond {
			continue
		}

		err = writeResponse(stdout, resp)
		if err != nil {
			return fmt.Errorf("write response: %w", err)
		}
	}
}

const headerSplitParts = 2

func handleRequest(ctx context.Context, req rpcRequest, client timesheetCreator) (rpcResponse, bool) {
	if req.ID == nil {
		return rpcResponse{}, false
	}

	resp := rpcResponse{JSONRPC: "2.0", ID: req.ID}
	switch req.Method {
	case "initialize":
		resp.Result = map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"serverInfo": map[string]any{
				"name":    "bexio-mcp",
				"version": "0.0.0",
			},
		}
		return resp, true
	case "tools/call":
		var params rpcToolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &rpcError{Code: -32602, Message: "invalid params"}
			return resp, true
		}
		if params.Name != "create_timesheet" {
			resp.Error = &rpcError{Code: -32601, Message: "method not found"}
			return resp, true
		}

		created, err := client.CreateTimesheet(ctx, params.Arguments)
		if err != nil {
			resp.Error = &rpcError{Code: -32603, Message: err.Error()}
			return resp, true
		}

		resp.Result = rpcToolCallResult{StructuredContent: created}
		return resp, true
	default:
		resp.Error = &rpcError{Code: -32601, Message: "method not found"}
		return resp, true
	}
}

func readPayload(reader *bufio.Reader) ([]byte, error) {
	contentLength := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		parts := strings.SplitN(line, ":", headerSplitParts)
		if len(parts) != headerSplitParts {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(parts[0]), "Content-Length") {
			continue
		}

		length, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("parse content length: %w", err)
		}
		contentLength = length
	}

	if contentLength < 0 {
		return nil, fmt.Errorf("invalid content length: %d", contentLength)
	}

	payload := make([]byte, contentLength)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, fmt.Errorf("read payload bytes: %w", err)
	}
	return payload, nil
}

func writeResponse(writer io.Writer, resp rpcResponse) error {
	payload, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}

	var header bytes.Buffer
	_, err = fmt.Fprintf(&header, "Content-Length: %d\r\n\r\n", len(payload))
	if err != nil {
		return fmt.Errorf("build header: %w", err)
	}

	_, err = writer.Write(header.Bytes())
	if err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	_, err = writer.Write(payload)
	if err != nil {
		return fmt.Errorf("write payload: %w", err)
	}
	return nil
}
