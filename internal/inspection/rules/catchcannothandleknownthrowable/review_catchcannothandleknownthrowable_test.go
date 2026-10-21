package catchcannothandleknownthrowable

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewCatchCannotHandleKnownThrowableArguments(t *testing.T) {
	p := expansionCryptoDBProbe{id: "CatchCannotHandleKnownThrowable", kinds: []syntax.NodeKind{syntax.KCatch}, check: func(ctx *analysis.Context, n syntax.Node) {
		check(ctx, n, "finding")
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}, PHP: phpversion.PHP85})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"match", "<?php try { throw new Error(match ($c) { 1 => 'one' }); } catch (UnhandledMatchError $e) {}", 0},
		{"property", "<?php try { throw new Error($o->message); } catch (Exception $e) {}", 0},
		{"interpolation", "<?php try { throw new Error(\"failed: $o\"); } catch (Exception $e) {}", 0},
		{"concatenation", "<?php try { throw new Error('failed: ' . $o); } catch (Exception $e) {}", 0},
		{"cast", "<?php try { throw new Error((string) $o); } catch (Exception $e) {}", 0},
		{"offset", "<?php try { throw new Error($o[0]); } catch (Exception $e) {}", 0},
		{"clone", "<?php try { throw new Error('x', 0, clone $o); } catch (Exception $e) {}", 0},
		{"class constant", "<?php try { throw new Error(Foo::BAR); } catch (Exception $e) {}", 0},
		{"global constant", "<?php try { throw new Error(BAR); } catch (Exception $e) {}", 0},
		{"spread", "<?php try { throw new Error(...$args); } catch (Exception $e) {}", 0},
		{"variable variable", "<?php try { throw new Error($$name); } catch (Exception $e) {}", 0},
		{"first-class callable", "<?php try { throw new Error(...); } catch (Exception $e) {}", 0},
		{"division", "<?php try { throw new Error('x', 1 / $d); } catch (Exception $e) {}", 0},
		{"no parentheses", "<?php try { throw new Error; } catch (Exception $e) {}", 1},
		{"no argument", "<?php try { throw new Error(); } catch (TypeError $e) {}", 1},
		{"string literal", "<?php try { throw new Error(('failed')); } catch (TypeError $e) {}", 1},
		{"variable", "<?php try { throw new Error($message); } catch (Exception $e) {}", 1},
		{"literals", "<?php try { throw new Error('failed', 3, null); } catch (Exception $e) {}", 1},
		{"magic constant", "<?php try { throw new Error(__FILE__); } catch (Exception $e) {}", 1},
		{"named literal", "<?php try { throw new Error(message: 'failed'); } catch (Exception $e) {}", 1},
		{"variable may be rejected", "<?php try { throw new Error($message); } catch (TypeError $e) {}", 0},
		{"integer may be rejected", "<?php try { throw new Error(3); } catch (TypeError $e) {}", 0},
		{"extra literal may be rejected", "<?php try { throw new Error('a', 'b'); } catch (TypeError $e) {}", 0},
		{"unknown name may be rejected", "<?php try { throw new Error(text: 'failed'); } catch (TypeError $e) {}", 0},
		{"argument count", "<?php try { throw new Error('a', 1, null, 4); } catch (ArgumentCountError $e) {}", 0},
		{"unrelated error type", "<?php try { throw new Error($message); } catch (ValueError $e) {}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opt := syntax.Options{Version: phpversion.PHP85}
			f := syntax.ParseBest("review.php", []byte(tc.src), opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}

type expansionCryptoDBProbe struct {
	id    string
	kinds []syntax.NodeKind
	check func(*analysis.Context, syntax.Node)
}

func (p expansionCryptoDBProbe) ID() string { return p.id }

func (p expansionCryptoDBProbe) Semantic() {}

func (p expansionCryptoDBProbe) Flow() {}

func (p expansionCryptoDBProbe) Kinds() []syntax.NodeKind { return p.kinds }

func (p expansionCryptoDBProbe) Check(ctx *analysis.Context, n syntax.Node) { p.check(ctx, n) }
