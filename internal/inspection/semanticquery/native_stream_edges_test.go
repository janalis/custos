package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

type nativeStreamProbe struct {
	id    string
	check func(*analysis.Context, syntax.Node, string)
}

func (p nativeStreamProbe) ID() string { return p.id }
func (nativeStreamProbe) Semantic()    {}
func (nativeStreamProbe) Flow()        {}
func (nativeStreamProbe) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KForeach, syntax.KReturn}
}
func (p nativeStreamProbe) Check(ctx *analysis.Context, n syntax.Node) { p.check(ctx, n, "proof") }

func TestNativeStreamUnknownContracts(t *testing.T) {
	for _, tc := range []struct {
		id, source string
		check      func(*analysis.Context, syntax.Node, string)
	}{
		{"DirectoryIteratorDotEntries", "$item=new SplFileInfo('file');foreach(new DirectoryIterator('.') as $bag['entry']){unlink($item->getPathname());}", CheckDirectoryIteratorDotEntries},
		{"FileLockFailureUnchecked", "fwrite();", CheckFileLockFailureUnchecked},
		{"FileTruncatedBeforeLock", "$box->stream=fopen('file','w');", CheckFileTruncatedBeforeLock},
		{"OwnedStreamNotClosed", "function load($box){$box->stream=fopen('file','r');return true;}", CheckOwnedStreamNotClosed},
		{"ReadModifyWriteLockTooLate", "file_put_contents(filename:'file',flags:LOCK_EX);", CheckReadModifyWriteLockTooLate},
		{"ReadModifyWriteLockTooLate", "file_put_contents('file',['entry'=>'value'],LOCK_EX);", CheckReadModifyWriteLockTooLate},
		{"StreamOpenFailureUnchecked", "fread();", CheckStreamOpenFailureUnchecked},
		{"StreamUseAfterClose", "fread();", CheckStreamUseAfterClose},
	} {
		t.Run(tc.id+"/"+tc.source, func(t *testing.T) {
			probe := nativeStreamProbe{tc.id, tc.check}
			engine, err := analysis.NewEngine([]analysis.Rule{probe}, analysis.Config{Only: []string{tc.id}})
			if err != nil {
				t.Fatal(err)
			}
			file := syntax.Parse("test.php", []byte("<?php "+tc.source), syntax.Options{})
			if len(file.Errors) != 0 {
				t.Fatalf("parse: %+v", file.Errors)
			}
			if findings := engine.Analyze(file); len(findings) != 0 {
				t.Fatalf("unknown facts must not establish a violation: %+v", findings)
			}
		})
	}
}

func TestNativeDirectoryUnrelatedCondition(t *testing.T) {
	probe := nativeStreamProbe{"DirectoryIteratorDotEntries", CheckDirectoryIteratorDotEntries}
	engine, err := analysis.NewEngine([]analysis.Rule{probe}, analysis.Config{Only: []string{probe.ID()}})
	if err != nil {
		t.Fatal(err)
	}
	file := syntax.Parse("test.php", []byte("<?php foreach(new DirectoryIterator('.') as $entry){if($skip){echo 'other';}unlink($entry->getPathname());}"), syntax.Options{})
	findings := engine.Analyze(file)
	if len(findings) != 1 || findings[0].Rule != probe.ID() {
		t.Fatalf("an unrelated condition does not exclude dot entries: %+v", findings)
	}
}
