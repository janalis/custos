package index

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/phpver"
	"custos/internal/syntax"
)

func TestDeclarationBuiltinResolution(t *testing.T) {
	for _, tt := range []struct {
		name, src string
		want      bool
	}{
		{"global", `define('FLAG', 7); class_alias('Original', 'Alias');`, true},
		{"namespace fallback", `namespace App; define('FLAG', 7); class_alias('Original', 'Alias');`, true},
		{"builtin imports", `namespace App; use function define as register; use function class_alias as rename; register('FLAG', 7); rename('Original', 'Alias');`, true},
		{"mixed case", `namespace App; use function DEFINE as register; use function CLASS_ALIAS as rename; REGISTER('FLAG', 7); RENAME('Original', 'Alias');`, true},
		{"shadow before", `namespace App; function define() {} function class_alias() {} define('FLAG', 7); class_alias('Original', 'Alias');`, false},
		{"shadow after", `namespace App; define('FLAG', 7); class_alias('Original', 'Alias'); function DEFINE() {} function CLASS_ALIAS() {}`, false},
		{"explicit globals", `namespace App; function define() {} function class_alias() {} \define('FLAG', 7); \class_alias('Original', 'Alias');`, true},
		{"imported user functions", `namespace App; use function Other\define; use function Other\class_alias; define('FLAG', 7); class_alias('Original', 'Alias');`, false},
		{"relative user functions", `namespace App; namespace\define('FLAG', 7); namespace\class_alias('Original', 'Alias');`, false},
		{"qualified user functions", `App\define('FLAG', 7); App\class_alias('Original', 'Alias');`, false},
		{"dynamic", `$define('FLAG', 7); $alias('Original', 'Alias');`, false},
		{"acquisition", `define(...); class_alias(...);`, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fs := extract(t, "calls.php", "<?php "+tt.src)
			if got := len(fs.Constants) == 1 && len(fs.ClassAliases) == 1; got != tt.want {
				t.Fatalf("constants=%+v aliases=%+v, want declarations=%v", fs.Constants, fs.ClassAliases, tt.want)
			}
			if !tt.want && len(fs.Constants)+len(fs.ClassAliases) != 0 {
				t.Fatal("partially indexed non-builtin calls")
			}
		})
	}
}

func TestDeclarationConstantsRetainSourceOrder(t *testing.T) {
	fs := extract(t, "order.php", `<?php define('FLAG', 1); const FLAG = 2; define('LAST', 3);`)
	ix := New(nil)
	ix.Add(fs)
	if c := ix.Constant("FLAG", 0); c == nil || c.Value != "1" {
		t.Fatalf("first declaration lost: %+v", c)
	}
	if len(fs.Constants) != 3 || fs.Constants[2].FQN != "LAST" {
		t.Fatalf("constant order: %+v", fs.Constants)
	}
}

func TestDeclarationBuiltinArguments(t *testing.T) {
	for _, tt := range []struct {
		name, src string
		want      bool
	}{
		{"positional optional", `define('FLAG', 7, false); class_alias('Original', 'Alias', false);`, true},
		{"named reordered", `define(value: 7, constant_name: 'FLAG'); class_alias(alias: 'Alias', class: 'Original');`, true},
		{"mixed", `define('FLAG', value: 7); class_alias('Original', alias: 'Alias');`, true},
		{"named optional first", `define(case_insensitive: false, value: 7, constant_name: 'FLAG'); class_alias(autoload: false, alias: 'Alias', class: 'Original');`, true},
		{"parentheses", `define(('FLAG'), 7); class_alias(('Original'), ('Alias'));`, true},
		{"missing required", `define(value: 7); class_alias(alias: 'Alias');`, false},
		{"duplicate", `define('FLAG', 7, value: 8); class_alias('Original', 'Alias', alias: 'Other');`, false},
		{"unknown name", `define('FLAG', 7, bad: false); class_alias('Original', 'Alias', bad: false);`, false},
		{"case sensitive names", `define(Constant_name: 'FLAG', value: 7); class_alias(Class: 'Original', alias: 'Alias');`, false},
		{"positional after named", `define(constant_name: 'FLAG', 7); class_alias(class: 'Original', 'Alias');`, false},
		{"too many", `define('FLAG', 7, false, 1); class_alias('Original', 'Alias', false, 1);`, false},
		{"unpacked optional", `define('FLAG', 7, ...$extra); class_alias('Original', 'Alias', ...$extra);`, false},
		{"unpacked required", `define(...$args); class_alias(...$args);`, false},
		{"nonliteral", `define($name, 7); class_alias($name, 'Alias');`, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fs := extract(t, "args.php", "<?php "+tt.src)
			if !tt.want {
				if len(fs.Constants)+len(fs.ClassAliases) != 0 {
					t.Fatalf("unexpected declarations: %+v %+v", fs.Constants, fs.ClassAliases)
				}
				return
			}
			if len(fs.Constants) != 1 || fs.Constants[0].FQN != "FLAG" || fs.Constants[0].Value != "7" || len(fs.ClassAliases) != 1 || fs.ClassAliases[0] != [2]string{"Alias", "Original"} {
				t.Fatalf("incorrect declarations: %+v %+v", fs.Constants, fs.ClassAliases)
			}
		})
	}
}

func TestDeclarationAliasLiteralNames(t *testing.T) {
	fs := extract(t, "aliases.php", `<?php
namespace App;
use Other\Original;
class_alias(Original::class, \Alias::CLASS);
class_alias(Original::OTHER, 'noConstant');
class_alias($class::class, 'noDynamicClass');
class_alias(Original::$member, 'noDynamicMember');
class_alias(self::class, 'noSelf');
class_alias(1, 'noNumber');
class_alias('', 'noEmpty');
`)
	if len(fs.ClassAliases) != 1 || fs.ClassAliases[0] != [2]string{"Alias", `Other\Original`} {
		t.Fatalf("aliases: %+v", fs.ClassAliases)
	}
}

func TestDeclarationNamedArgumentVersions(t *testing.T) {
	for _, ver := range []phpver.Version{0, phpver.PHP74, phpver.PHP80, phpver.PHP85} {
		f := syntax.Parse("version.php", []byte(`<?php define(value: 7, constant_name: 'FLAG'); class_alias(alias: 'Alias', class: 'Original');`), syntax.Options{Version: phpver.PHP85})
		f.Version = ver
		fs := Extract(f)
		want := ver == 0 || ver >= phpver.PHP80
		if (len(fs.Constants) == 1 && len(fs.ClassAliases) == 1) != want {
			t.Fatalf("version %v: %+v", ver, fs)
		}
	}
}

func TestDeclarationMalformedAST(t *testing.T) {
	x := &extractor{f: &syntax.File{}, out: &FileSymbols{}}
	for _, call := range []*syntax.FuncCall{
		{},
		{Args: &syntax.ArgList{Args: []syntax.Expr{&syntax.Arg{}}}},
		{Args: &syntax.ArgList{Args: []syntax.Expr{&syntax.Arg{Value: &syntax.Literal{}, ByRef: true}}}},
	} {
		x.define(call)
		x.classAlias(call)
	}
	if len(x.out.Constants)+len(x.out.ClassAliases) != 0 {
		t.Fatal("malformed AST declared symbols")
	}
}

func BenchmarkExtractDeclarationCalls(b *testing.B) {
	for _, calls := range []bool{false, true} {
		b.Run(fmt.Sprintf("calls=%v", calls), func(b *testing.B) {
			var src strings.Builder
			src.WriteString("<?php namespace App; use function define as register;\n")
			for i := range 1000 {
				fmt.Fprintf(&src, "function f%d() {}\n", i)
				if calls {
					fmt.Fprintf(&src, "register(value: %d, constant_name: 'FLAG%d');\n", i, i)
				}
			}
			f := syntax.Parse("bench.php", []byte(src.String()), syntax.Options{Version: phpver.PHP85})
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				Extract(f)
			}
		})
	}
}
