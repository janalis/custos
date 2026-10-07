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
	codeInvalidRequest = -32600
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
	return &conn{r: bufio.NewReaderSize(r, maxHeaderLine), w: w}
}

// maxMessageSize bounds a JSON-RPC message body (a variable for tests). A
// document of syntax.MaxFileSize bytes stays well below it even fully
// JSON-escaped; a larger Content-Length is answered with an error and its
// body skipped, instead of allocating whatever the header claims (a huge
// value used to crash the server with an out-of-memory error).
var maxMessageSize = 128 << 20

// maxHeaderLine bounds one header line (the read buffer size).
const maxHeaderLine = 64 << 10

func (c *conn) read() (*message, error) {
	length := -1
	for {
		raw, err := c.r.ReadSlice('\n')
		if err == bufio.ErrBufferFull {
			return nil, fmt.Errorf("lsp: header line longer than %d bytes", maxHeaderLine)
		}
		if err != nil {
			return nil, err
		}
		line := strings.TrimRight(string(raw), "\r\n")
		if line == "" {
			break
		}
		name, value, ok := strings.Cut(line, ":")
		if ok && strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			length, err = strconv.Atoi(strings.TrimSpace(value))
			if err != nil || length < 0 {
				return nil, fmt.Errorf("lsp: bad Content-Length %q", value)
			}
		}
	}
	if length < 0 {
		return nil, errors.New("lsp: missing Content-Length")
	}
	if length > maxMessageSize {
		if _, err := io.CopyN(io.Discard, c.r, int64(length)); err != nil {
			return nil, err
		}
		return nil, &rpcError{Code: codeInvalidRequest, Message: fmt.Sprintf("message of %d bytes exceeds the %d byte limit", length, maxMessageSize)}
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
	// Every raw field is valid JSON (decoded from the client or produced
	// by mustJSON), so marshalling cannot fail.
	body := mustJSON(m)
	frame := append(fmt.Appendf(nil, "Content-Length: %d\r\n\r\n", len(body)), body...)
	c.wmu.Lock()
	defer c.wmu.Unlock()
	_, err := c.w.Write(frame)
	return err
}

func (c *conn) reply(id *json.RawMessage, result any, rerr *rpcError) error {
	m := &message{ID: id, Error: rerr}
	if rerr == nil {
		m.Result = mustJSON(result)
	}
	return c.write(m)
}

func (c *conn) notify(method string, params any) error {
	return c.write(&message{Method: method, Params: mustJSON(params)})
}

// request sends a server->client request; the response is ignored.
func (c *conn) request(method string, params any) error {
	id := json.RawMessage(strconv.FormatInt(c.nextID.Add(1), 10))
	return c.write(&message{ID: &id, Method: method, Params: mustJSON(params)})
}
