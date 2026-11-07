package analysis

import (
	"testing"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

func TestUnchangedFileReadsTheProjectIndex(t *testing.T) {
	src := []byte(`<?php class Held { function run() {} }`)
	indexed := syntax.Parse("held.php", src, syntax.Options{})
	fs := index.Extract(indexed)
	fs.Digest = index.Digest(indexed)
	project := index.New(nil)
	project.Add(fs)
	project.Freeze()
	e, err := NewEngine(nil, Config{})
	if err != nil {
		t.Fatal(err)
	}
	e = e.WithIndex(project)
	if ctx := NewTestContext(e, syntax.Parse("held.php", src, syntax.Options{})); ctx.Index() != project {
		t.Fatal("an unchanged file got a layer of its own")
	}
	edited := syntax.Parse("held.php", []byte(`<?php class Held { function stop() {} }`), syntax.Options{})
	ctx := NewTestContext(e, edited)
	if ix := ctx.Index(); ix == project || ix.FindMethod("Held", "stop", 0) == nil || ix.FindMethod("Held", "run", 0) != nil {
		t.Fatal("an edited file read the symbols indexed before the edit")
	}
	if ctx := NewTestContext(e, syntax.Parse("other.php", src, syntax.Options{})); ctx.Index() == project {
		t.Fatal("a file the project does not index read it as its own")
	}
}
