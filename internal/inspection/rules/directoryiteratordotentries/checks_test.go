package directoryiteratordotentries

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"DirectoryIteratorDotEntries"}})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

// TestChecks counts the findings on short sources.
func TestChecks(t *testing.T) {
	e := engine(t)
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"foreach(new DirectoryIterator('/tmp') as $entry){$entry=new SplFileInfo('/safe/file');unlink($entry->getPathname());}", 0},
		{"foreach(new DirectoryIterator('/tmp') as $entry){if($flag){$entry=new SplFileInfo('/safe/file');}unlink($entry->getPathname());}", 0},
		{"foreach(new DirectoryIterator('/tmp') as $entry){$f=function(){$entry=1;};unlink($entry->getPathname());}", 1},
		{"foreach(new DirectoryIterator('/tmp') as $entry){unset($entry);unlink($entry->getPathname());}", 0},
		{"foreach(new DirectoryIterator('/tmp') as $entry){$entry++;unlink($entry->getPathname());}", 0},
		{"foreach(new DirectoryIterator('/tmp') as $entry){foreach($other as $entry){}unlink($entry->getPathname());}", 0},
		{"harmless(); $v=1+2; ++$v; echo ''; foreach([] as $entry){} function benign(){return 1;} $object->unrelated();", 0},
		{"foreach(new DirectoryIterator('/tmp') as $entry){if($other){continue;}unlink($entry->getPathname());}", 1},
		{"unlink('/safe'); unlink(makePath()->getPathname());", 0},
		{"foreach($iterator as $entry){unlink($entry->getPathname());}", 0},
		{"foreach(new ArrayIterator([]) as $entry){unlink($entry->getPathname());}", 0},
		{"foreach(new DirectoryIterator('/tmp') as $entry){if($enabled){unlink($entry->getPathname());}}", 1},
		{"foreach(new DirectoryIterator('/tmp') as $entry){if(!isAllowed()){unlink($entry->getPathname());}}", 1},
		{"foreach(new DirectoryIterator('/tmp') as $entry){if($other->isDot()){continue;}unlink($entry->getPathname());}", 1},
		{"$item=new SplFileInfo('file');foreach(new DirectoryIterator('.') as $bag['entry']){unlink($item->getPathname());}", 0},
		{"foreach(new DirectoryIterator('.') as $entry){if($skip){echo 'other';}unlink($entry->getPathname());}", 1},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
