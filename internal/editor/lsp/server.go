package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/project/config"
	"custos/internal/semantic/index"
)

// Debounce is the delay between the last edit and re-analysis.
var Debounce = 150 * time.Millisecond

const (
	cmdFixFile = "custos.fixFile"
	cmdFixRule = "custos.fixRule"
	kindFixAll = "source.fixAll.custos"
)

type document struct {
	uri      string
	path     string
	version  int
	text     []byte
	lines    *syntax.LineIndex
	findings []diagnostic.Finding // for analyzedVersion
	analyzed int                  // version the findings belong to (-1 = none)
	timer    *time.Timer
}

// Server is the custos language server.
type Server struct {
	c        *conn
	mu       sync.Mutex
	docs     map[string]*document
	engine   *analysis.Engine
	cfg      *config.Config
	initOpts json.RawMessage
	root     string
	resolve  bool // client supports codeAction/resolve
	// watchDynamic: client supports dynamic registration of file watchers.
	watchDynamic bool
	// indexing is true while the initial project index is being built;
	// file changes arriving meanwhile are queued in pending and replayed.
	indexing bool
	pending  []fileChange
	index    *index.Index
	shutdown bool
	// sem bounds concurrent analyses.
	sem chan struct{}
}

// Serve runs the server on r/w until exit.
func Serve(ctx context.Context, r io.Reader, w io.Writer) error {
	s := &Server{c: newConn(r, w), docs: map[string]*document{}, sem: make(chan struct{}, runtime.GOMAXPROCS(0))}
	for {
		m, err := s.c.read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			var rerr *rpcError
			if errors.As(err, &rerr) {
				_ = s.c.reply(nil, nil, rerr)
				continue
			}
			return err
		}
		if m.Method == "" {
			continue // response to one of our requests
		}
		if m.Method == "exit" {
			return nil
		}
		result, rerr := s.handle(m)
		if m.ID != nil {
			if err := s.c.reply(m.ID, result, rerr); err != nil {
				return err
			}
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

// guard, deferred, recovers a panic: it is logged and onPanic (if any) runs.
func (s *Server) guard(what string, onPanic func()) {
	if r := recover(); r != nil {
		s.logf("%s panicked: %v", what, r)
		if onPanic != nil {
			onPanic()
		}
	}
}

func (s *Server) logf(format string, args ...any) {
	_ = s.c.notify("window/logMessage", map[string]any{"type": 3, "message": "custos: " + fmt.Sprintf(format, args...)})
}
