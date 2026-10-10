package untrustedfilesystempath

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestServerGeneratedUploadPath(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"UntrustedFilesystemPath"}})
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("upload.php", []byte("<?php $content=file_get_contents($_FILES['file']['tmp_name']);"), syntax.Options{})
	if got := e.Analyze(f); len(got) != 0 {
		t.Fatalf("server generated upload path must not be tainted: %+v", got)
	}
}
