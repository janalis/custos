package names

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"

	"custos/internal/syntax"
)

func TestAnonymousSymbolIdentity(t *testing.T) {
	f := syntax.Parse("/project/classes.php", []byte(`<?php namespace App; class Named {} $x = new class {}; $y = new class {};`), syntax.Options{})
	r := New(f)
	var classes []*syntax.ClassLike
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.ClassLike); ok {
			classes = append(classes, c)
		}
		return true
	})
	if len(classes) != 3 {
		t.Fatalf("classes: %d", len(classes))
	}
	if r.SymbolFQN(nil) != "" || r.SymbolFQN(classes[0]) != `App\Named` {
		t.Fatal("named identity")
	}
	first := r.SymbolFQN(classes[1])
	want := fmt.Sprintf("@anonymous:%x:%d", sha256.Sum256([]byte(f.Path)), classes[1].Span().Start)
	if first != want || first != New(f).SymbolFQN(classes[1]) || first == r.SymbolFQN(classes[2]) {
		t.Fatal("unstable identity", first)
	}
	if r.DeclFQN(classes[1]) != "" {
		t.Fatal("anonymous declaration became named")
	}
	f.Path = "/other/classes.php"
	if first == New(f).SymbolFQN(classes[1]) {
		t.Fatal("paths collide")
	}
	for _, name := range []string{first, `\` + first} {
		if !IsAnonymousClassName(name) {
			t.Fatal("missing anonymous marker", name)
		}
	}
	for _, invalid := range []string{`App\Named`, "@anonymous:", "@anonymous:" + strings.Repeat("a", 64) + ":", "@anonymous:" + strings.Repeat("g", 64) + ":1", "@anonymous:" + strings.Repeat("a", 64) + ":-1", "@anonymous:" + strings.Repeat("a", 64) + ":4294967296"} {
		if IsAnonymousClassName(invalid) {
			t.Fatal("invalid anonymous marker", invalid)
		}
	}
}

func BenchmarkAnonymousSymbolIdentity(b *testing.B) {
	f := syntax.Parse("/project/classes.php", []byte(`<?php $x = new class {};`), syntax.Options{})
	var c *syntax.ClassLike
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if cls, ok := n.(*syntax.ClassLike); ok {
			c = cls
		}
		return true
	})
	r := New(f)
	b.ReportAllocs()
	for b.Loop() {
		_ = r.SymbolFQN(c)
	}
}
