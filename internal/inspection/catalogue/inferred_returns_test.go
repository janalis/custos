package catalogue

import (
	"os"
	"path/filepath"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/project"
	"custos/internal/semantic/infer"
)

// The return types the project index infers from bodies in other files
// count for ReturnTypeCanBeDeclared's hierarchy check: a parent returning a
// value forbids `void`, one falling off its end (inferred null) does not.
func TestReturnTypeHierarchyUsesInferredReturns(t *testing.T) {
	dir := t.TempDir()
	write := func(name, src string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(src), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	files := []string{
		write("parent.php", `<?php
abstract class Connection {
    public function connect() { return true; }
    public function close() { $this->flush(); }
    protected function flush(): void {}
}
`),
		write("child.php", `<?php
final class Stream extends Connection {
    public function connect() { throw new \LogicException('offline'); }
    public function close() { $this->flush(); }
}
`),
	}
	var rules []analysis.Rule
	for _, r := range All() {
		if r.ID() == "ReturnTypeCanBeDeclared" {
			rules = append(rules, r)
		}
	}
	e, err := analysis.NewEngine(rules, analysis.Config{PHP: phpversion.PHP84, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	opt := syntax.Options{Version: phpversion.PHP84}
	e = e.WithIndex(project.BuildIndex(files, opt, infer.Options{}))
	var got []string
	for _, r := range project.Run(e, files, opt) {
		if filepath.Base(r.Path) != "child.php" {
			continue
		}
		for _, f := range r.Findings {
			got = append(got, f.Message)
		}
	}
	if len(got) != 1 || got[0] != "Declare ': void' as the return type." {
		t.Fatalf("child findings: got %q, want only close() as void", got)
	}
}
