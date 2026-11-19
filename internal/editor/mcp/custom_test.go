package mcp

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/inspection/meta"
	"github.com/janalis/custos/internal/inspection/pattern"
)

func TestCustomRules(t *testing.T) {
	s, ctx, root, _ := agentFixture(t)
	writeFile(t, filepath.Join(root, "custos.json"), `{"php":"8.2","customRules":[{"id":"NoTriple","match":"!!!$X","message":"Simplify !!!$X","fix":"!$X"}],"customRuleFiles":["./rules.json"]}`)
	writeFile(t, filepath.Join(root, "rules.json"), `[{"id":"NoMore","match":"$more","message":"m"}]`)
	cp := invoke(ctx, t, s, "analyze", Input{})["checkpointId"].(string)
	found := invoke(ctx, t, s, "findings", Input{CheckpointID: cp, Rule: "notriple"})
	if found["total"] != 3 || !strings.Contains(fmt.Sprint(found["items"]), "Simplify !!!$token") {
		t.Fatal(found)
	}
	if out := invoke(ctx, t, s, "findings", Input{CheckpointID: cp, Rule: "NoMore"}); out["total"] != 1 {
		t.Fatal(out)
	}
	info := invoke(ctx, t, s, "rules", Input{Rule: "NoTriple"})
	item := info["items"].([]any)[0].(Output)
	if info["total"] != 1 || len(item["suppressions"].([]string)) != 1 || !strings.Contains(item["explanation"].(meta.Description).Summary, "Simplify") {
		t.Fatal(info)
	}
	if out := invoke(ctx, t, s, "prepare_fixes", Input{CheckpointID: cp, Rule: "NoTriple", Iterative: true}); out["planId"] == nil {
		t.Fatal(out)
	}
	// Rules the program adds are known too, even when the configuration
	// does not load.
	extra, err := pattern.Compile(pattern.Definition{ID: "FromGo", Match: pattern.Patterns{"$more"}, Message: "go"})
	if err != nil {
		t.Fatal(err)
	}
	s.rules = append(s.rules, extra)
	if out := invoke(ctx, t, s, "rules", Input{Rule: "fromgo"}); out["total"] != 1 {
		t.Fatal(out)
	}
	if got := s.customRules(nil); len(got) != 1 || got[0].ID() != "FromGo" {
		t.Fatal(got)
	}
	// A checkpoint keeps the rules it was analyzed with: once its snapshot
	// is gone, the workspace's configuration as it is now decides.
	writeFile(t, filepath.Join(root, "custos.json"), `{"customRules":[{"id":"Bad"}]}`)
	if _, err := s.call(ctx, "findings", Input{CheckpointID: cp, Rule: "NoMore"}); err == nil || !strings.Contains(err.Error(), "inputs changed (custos.json)") {
		t.Fatal(err)
	}
	delete(s.snapshots, cp)
	if cfg := s.checkpointConfig(s.workspaces["default"], s.checkpoints[cp]); cfg != nil {
		t.Fatal(cfg)
	}
}
