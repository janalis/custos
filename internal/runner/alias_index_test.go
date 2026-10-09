package runner

import (
	"path/filepath"
	"testing"

	"custos/internal/phpver"
	"custos/internal/syntax"
)

func TestBuildIndexAliasOnlyFile(t *testing.T) {
	dir := t.TempDir()
	original := filepath.Join(dir, "original.php")
	alias := filepath.Join(dir, "alias.php")
	writeFile(t, original, `<?php namespace App; final class Product { public function count(): int { return 1; } }`)
	writeFile(t, alias, `<?php use function class_alias as register; register(alias: 'ProductAlias', class: \App\Product::class);`)
	for _, files := range [][]string{{original, alias}, {alias, original}} {
		ix := BuildIndex(files, syntax.Options{Version: phpver.PHP85})
		if got := ix.Class("ProductAlias", phpver.PHP85); got == nil || got.FQN != `App\Product` {
			t.Fatalf("alias-only file missing: %+v", got)
		}
		if method := ix.FindMethod("ProductAlias", "count", phpver.PHP85); method == nil || method.Return != "int" {
			t.Fatalf("aliased method missing: %+v", method)
		}
		ix.Remove(alias)
		if ix.Class("ProductAlias", phpver.PHP85) != nil || ix.Class(`App\Product`, phpver.PHP85) == nil {
			t.Fatal("removing alias file did not preserve the original class")
		}
	}
}

func TestBuildIndexDeclarationShadows(t *testing.T) {
	dir := t.TempDir()
	calls := filepath.Join(dir, "calls.php")
	shadow := filepath.Join(dir, "shadow.php")
	writeFile(t, calls, `<?php namespace App; class Original {} define('FLAG', 7); class_alias(Original::class, 'Alias');`)
	writeFile(t, shadow, `<?php namespace App; function define() {} function class_alias() {}`)
	for _, files := range [][]string{{calls, shadow}, {shadow, calls}} {
		ix := BuildIndex(files, syntax.Options{Version: phpver.PHP85})
		if ix.Constant("FLAG", phpver.PHP85) != nil || ix.Class("Alias", phpver.PHP85) != nil {
			t.Fatal("project functions did not suppress declarations")
		}
		ix.Remove(shadow)
		if ix.Constant("FLAG", phpver.PHP85) == nil || ix.Class("Alias", phpver.PHP85) == nil {
			t.Fatal("removing project functions did not restore declarations")
		}
	}
}
