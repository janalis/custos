// Package lsp implements the custos language server (LSP over stdio):
// push diagnostics, quick-fix code actions and fix-all commands.
package lsp

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// message is a JSON-RPC 2.0 request, notification or response.
type message struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method,omitempty"`
	Params  json.RawMessage  `json:"params,omitempty"`
	Result  json.RawMessage  `json:"result,omitempty"`
	Error   *rpcError        `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *rpcError) Error() string { return fmt.Sprintf("jsonrpc %d: %s", e.Code, e.Message) }

const (
	codeParseError     = -32700
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
	codeInternalError  = -32603
	codeRequestFailed  = -32803
)

// conn frames JSON-RPC messages with Content-Length headers.
type conn struct {
	r      *bufio.Reader
	w      io.Writer
	wmu    sync.Mutex
	nextID atomic.Int64
}

func newConn(r io.Reader, w io.Writer) *conn {
	return &conn{r: bufio.NewReaderSize(r, 64<<10), w: w}
}

func (c *conn) read() (*message, error) {
	length := -1
	for {
		line, err := c.r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		name, value, ok := strings.Cut(line, ":")
		if ok && strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			length, err = strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				return nil, fmt.Errorf("lsp: bad Content-Length %q", value)
			}
		}
	}
	if length < 0 {
		return nil, errors.New("lsp: missing Content-Length")
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(c.r, body); err != nil {
		return nil, err
	}
	var m message
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, &rpcError{Code: codeParseError, Message: err.Error()}
	}
	return &m, nil
}

func (c *conn) write(m *message) error {
	m.JSONRPC = "2.0"
	body, err := json.Marshal(m)
	if err != nil {
		return err
	}
	c.wmu.Lock()
	defer c.wmu.Unlock()
	if _, err := fmt.Fprintf(c.w, "Content-Length: %d\r\n\r\n", len(body)); err != nil {
		return err
	}
	_, err = c.w.Write(body)
	return err
}

func (c *conn) reply(id *json.RawMessage, result any, rerr *rpcError) error {
	m := &message{ID: id, Error: rerr}
	if rerr == nil {
		b, err := json.Marshal(result)
		if err != nil {
			return err
		}
		m.Result = b
	}
	return c.write(m)
}

func (c *conn) notify(method string, params any) error {
	b, err := json.Marshal(params)
	if err != nil {
		return err
	}
	return c.write(&message{Method: method, Params: b})
}

// request sends a server->client request; the response is ignored.
func (c *conn) request(method string, params any) error {
	b, err := json.Marshal(params)
	if err != nil {
		return err
	}
	id := json.RawMessage(strconv.FormatInt(c.nextID.Add(1), 10))
	return c.write(&message{ID: &id, Method: method, Params: b})
}
