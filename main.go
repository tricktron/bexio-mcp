package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	JSONRPC string      `json:"jsonrpc"`
	ID      *int        `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
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
	_ = cfg

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
		if err := json.Unmarshal(payload, &req); err != nil {
			return fmt.Errorf("decode request: %w", err)
		}

		resp, shouldRespond := handleRequest(req)
		if !shouldRespond {
			continue
		}

		if err := writeResponse(stdout, resp); err != nil {
			return fmt.Errorf("write response: %w", err)
		}
	}
}

func handleRequest(req rpcRequest) (rpcResponse, bool) {
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
		resp.Result = map[string]any{}
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

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
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
		return nil, err
	}
	return payload, nil
}

func writeResponse(writer io.Writer, resp rpcResponse) error {
	payload, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}

	var header bytes.Buffer
	if _, err := fmt.Fprintf(&header, "Content-Length: %d\r\n\r\n", len(payload)); err != nil {
		return fmt.Errorf("build header: %w", err)
	}

	if _, err := writer.Write(header.Bytes()); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	if _, err := writer.Write(payload); err != nil {
		return fmt.Errorf("write payload: %w", err)
	}
	return nil
}
