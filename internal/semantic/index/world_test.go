package index_test

import (
	"testing"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/stubs"
)

func TestClassWord(t *testing.T) {
	for name, want := range map[string]string{"SplObjectStorage": "spl", "PDOStatement": "pdo", "PDO": "pdo", "DateTime": "date", "Zend_Db": "zend", "XML_Parser": "xml", "Yac": "yac", "stdClass": "std"} {
		if got := index.ClassWord(name); got != want {
			t.Errorf("ClassWord(%s) = %s, want %s", name, got, want)
		}
	}
}

func TestWorldLayers(t *testing.T) {
	project := index.New(stubs.Index())
	if project.World().Complete {
		t.Fatal("a fresh index knows nothing of its world")
	}
	project.SetWorld(index.World{Complete: true, Blind: []string{`app\tests\`}})
	project.Freeze()
	file := index.New(project)
	w := file.World()
	if !w.Complete || !w.Unseen(`\App\Tests\Case`) || w.Unseen(`App\Model`) {
		t.Fatalf("layered world: %+v", w)
	}
}

func TestDynamicAliases(t *testing.T) {
	for _, tc := range []struct {
		src     string
		dynamic bool
	}{
		{"<?php class_alias('A', 'B');", false},
		{"<?php class_alias($original, 'B');", true},
		{"<?php class_alias(...$pair);", true},
		{"<?php function f() {}", false},
	} {
		project := index.New(stubs.Index())
		project.Add(index.Extract(syntax.Parse("a.php", []byte(tc.src), syntax.Options{})))
		project.Freeze()
		// The frozen layer answers from its memo; a layer above asks it.
		if project.DynamicAliases() != tc.dynamic || index.New(project).DynamicAliases() != tc.dynamic {
			t.Fatalf("%s: want %v", tc.src, tc.dynamic)
		}
	}
}

func TestBuiltinPrefixes(t *testing.T) {
	p := index.New(stubs.Index()).BuiltinPrefixes()
	if !p.Words["str"] || !p.Words["apcu"] || p.Words["newext"] {
		t.Fatal("leading words of global builtin functions")
	}
	if !p.Namespaces["mongodb"] || !p.Namespaces["random"] || p.Namespaces["app"] {
		t.Fatal("namespaces of builtins")
	}
}
