package analysis

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	"custos/internal/semantic/names"
)

// usesNames reports the file after resolving names, which the test makes
// crash; usesTypes builds types on top of names.
type usesNames struct{ patternRule }

func (usesNames) CheckFile(ctx *Context) {
	ctx.Names()
	ctx.ReportNode(ctx.File, "names")
}

type usesTypes struct{ patternRule }

func (usesTypes) CheckFile(ctx *Context) {
	ctx.Types()
	ctx.ReportNode(ctx.File, "types")
}

type plainFileRule struct{ patternRule }

func (plainFileRule) CheckFile(ctx *Context) { ctx.ReportNode(ctx.File, "plain") }

func TestSharedDataFailureSkipsOnlyTheChecksNeedingIt(t *testing.T) {
	e, err := NewEngine([]Rule{usesNames{patternRule{"SecurityAdvisories"}}, usesTypes{patternRule{"NestedNotOperators"}}, plainFileRule{patternRule{"UnnecessarySemicolon"}}}, Config{EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	old := newResolver
	t.Cleanup(func() { newResolver = old })
	newResolver = func(*syntax.File) *names.Resolver { panic("resolver bug") }
	file := syntax.Parse("x.php", []byte("<?php ;"), syntax.Options{})
	result := e.AnalyzeResult(file)
	got := result.Findings
	if len(got) != 2 || got[0].Rule != "internal" || !strings.Contains(got[0].Message, "could not complete the name resolution of this file") || got[1].Message != "plain" {
		t.Fatalf("got %+v", got)
	}
	if len(result.Incomplete) != 1 || result.Incomplete[0].Reason != "internal-failure" {
		t.Fatalf("status %+v", result.Incomplete)
	}
	// Isolated checks still crash on their own: that rule is disabled and
	// the others run.
	e, err = NewEngine([]Rule{usesNames{patternRule{"SecurityAdvisories"}}, crashRule{}, okRule{}}, Config{EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	got = e.Analyze(syntax.Parse("x.php", []byte("<?php ;\n$a = !$b;"), syntax.Options{}))
	if len(got) != 3 || !strings.Contains(got[0].Message, "rule UnnecessarySemicolon crashed") || !strings.Contains(got[1].Message, "name resolution") || got[2].Rule != "NestedNotOperators" {
		t.Fatalf("got %+v", got)
	}
	// A test context records the failure as well.
	c := NewTestContext(e, file)
	for range 2 {
		func() {
			defer func() {
				if _, ok := recover().(envFailure); !ok {
					t.Fatal("expected a shared data failure")
				}
			}()
			c.Types()
		}()
	}
	if len(c.failed) != 2 {
		t.Fatalf("failed %v", c.failed)
	}
}
