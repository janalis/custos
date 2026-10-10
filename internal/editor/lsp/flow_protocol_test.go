package lsp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"custos/internal/inspection/rules/untrustedshellcommand"
	"custos/internal/testing/testbudget"
)

func TestOpenChangeAndCloseRefreshDependentFlowDiagnostics(t *testing.T) {
	useRules(t, untrustedshellcommand.New())
	dir := t.TempDir()
	wrapper := filepath.Join(dir, "wrapper.php")
	unsafe := "<?php function invoke($arg) { system($arg); }"
	safe := "<?php function invoke($arg) { return $arg; }"
	if err := os.WriteFile(wrapper, []byte(unsafe), 0o644); err != nil {
		t.Fatal(err)
	}
	cl := startServer(t)
	cl.initialize(map[string]any{"rootUri": "file://" + dir})
	if err := cl.c.notify("initialized", map[string]any{}); err != nil {
		t.Fatal(err)
	}
	cl.logContaining("indexed")
	mainURI := "file://" + filepath.Join(dir, "main.php")
	wrapperURI := "file://" + wrapper
	waitDependent := func(count int) {
		t.Helper()
		deadline := time.NewTimer(testbudget.Of(5 * time.Second))
		defer deadline.Stop()
		for {
			select {
			case m, ok := <-cl.out:
				if !ok {
					t.Fatal("server closed before dependent diagnostics")
				}
				if m.Method != "textDocument/publishDiagnostics" {
					continue
				}
				var p publishDiagnosticsParams
				if err := json.Unmarshal(m.Params, &p); err != nil {
					t.Fatal(err)
				}
				if p.URI == mainURI && len(p.Diagnostics) == count {
					for _, d := range p.Diagnostics {
						if d.Code != "UntrustedShellCommand" {
							t.Fatalf("unexpected dependent diagnostic: %+v", d)
						}
					}
					return
				}
			case <-deadline.C:
				t.Fatalf("dependent diagnostics did not refresh to %d findings", count)
			}
		}
	}
	cl.open(mainURI, "<?php invoke($_GET['command']);")
	waitDependent(1)
	// Opening an unsaved safe wrapper must immediately update the main
	// document, even though its own buffer version did not change.
	cl.open(wrapperURI, safe)
	waitDependent(0)
	cl.change(wrapperURI, 2, unsafe)
	waitDependent(1)
	cl.change(wrapperURI, 3, safe)
	waitDependent(0)
	// Closing the unsaved wrapper restores the saved disk contract.
	if err := cl.c.notify("textDocument/didClose", map[string]any{"textDocument": map[string]any{"uri": wrapperURI}}); err != nil {
		t.Fatal(err)
	}
	waitDependent(1)
	if result := cl.call(2, "shutdown", nil); result.Error != nil {
		t.Fatal(result.Error)
	}
}
