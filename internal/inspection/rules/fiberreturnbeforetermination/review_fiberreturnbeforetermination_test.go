package fiberreturnbeforetermination

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewFiberReturnBeforeTerminationReceiverUses(t *testing.T) {
	p := reviewProbe{id: "FiberReturnBeforeTermination", kinds: []syntax.NodeKind{syntax.KMethodCall}, check: func(ctx *analysis.Context, n syntax.Node) {
		check(ctx, n, "finding")
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}, PHP: phpversion.PHP85})
	if err != nil {
		t.Fatal(err)
	}
	const alloc = "$f = new Fiber(function () { Fiber::suspend(); return 1; }); $f->start(); "
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"top level", "<?php " + alloc + "$f->getReturn();", 1},
		{"function", "<?php function a() { " + alloc + "return $f->getReturn(); }", 1},
		{"arrow callback", "<?php function a() { $f = new Fiber(fn() => Fiber::suspend()); $f->start(); return $f->getReturn(); }", 1},
		{"later use", "<?php function a() { " + alloc + "$r = $f->getReturn(); $f->resume(); return $r; }", 1},
		{"parenthesized start", "<?php function a() { $f = new Fiber(fn() => Fiber::suspend()); ($f)->start(); return $f->getReturn(); }", 1},
		{"unrelated declarations", "<?php function other($f) { $f->resume(); } class K { public $p { set($f) { $f->resume(); } } function m($f) { $f->resume(); } } $g = function ($f) { $f->resume(); }; $h = function () use ($g) {}; " + alloc + "$f->getReturn();", 1},
		{"unrelated call", "<?php function a($g) { " + alloc + "$g->resume(); strlen('x'); $name(); return $f->getReturn(); }", 1},
		{"while resume", "<?php function a() { " + alloc + "while ($f->isSuspended()) { $f->resume(); } return $f->getReturn(); }", 0},
		{"do-while resume", "<?php function a() { " + alloc + "do { $f->resume(); } while (!$f->isTerminated()); return $f->getReturn(); }", 0},
		{"for resume", "<?php function a() { " + alloc + "for ($i = 0; $i < 1; $i++) { $f->resume(); } return $f->getReturn(); }", 0},
		{"if resume", "<?php function a() { " + alloc + "if (!$f->isTerminated()) { $f->resume(); } return $f->getReturn(); }", 0},
		{"try resume", "<?php function a() { " + alloc + "try { $f->resume(); } finally { } return $f->getReturn(); }", 0},
		{"straight-line resume", "<?php function a() { " + alloc + "$f->resume(); return $f->getReturn(); }", 0},
		{"throw", "<?php function a() { " + alloc + "if (rand(0, 1)) { $f->throw(new Exception()); } return $f->getReturn(); }", 0},
		{"guarded block", "<?php function a() { " + alloc + "if ($f->isTerminated()) { return $f->getReturn(); } return 0; }", 0},
		{"guarded ternary", "<?php function a() { " + alloc + "return $f->isTerminated() ? $f->getReturn() : null; }", 0},
		{"passed to call", "<?php function a() { " + alloc + "drive($f); return $f->getReturn(); }", 0},
		{"alias", "<?php function a() { " + alloc + "$g = $f; $g->resume(); return $f->getReturn(); }", 0},
		{"callable array", "<?php function a() { " + alloc + "$r = [$f, 'resume']; $r(); return $f->getReturn(); }", 0},
		{"arrow capture", "<?php function a() { " + alloc + "$t = fn() => $f->resume(); $t(); return $f->getReturn(); }", 0},
		{"closure capture", "<?php function a() { " + alloc + "$t = function () use ($f) { $f->resume(); }; $t(); return $f->getReturn(); }", 0},
		{"anonymous class argument", "<?php function a() { " + alloc + "new class($f) { function __construct($x) { $x->resume(); } }; return $f->getReturn(); }", 0},
		{"variable variable", "<?php function a($n) { " + alloc + "$$n->resume(); return $f->getReturn(); }", 0},
		{"compact", "<?php function a() { " + alloc + "drive(\\compact('f')); return $f->getReturn(); }", 0},
		{"defined variables", "<?php function a() { " + alloc + "drive(get_defined_vars()); return $f->getReturn(); }", 0},
		{"reassigned before", "<?php function a($g) { $f = $g; $f = new Fiber(fn() => Fiber::suspend()); $f->start(); return $f->getReturn(); }", 0},
		{"loop repeats later resume", "<?php function a() { " + alloc + "for ($i = 0; $i < 2; $i++) { if ($i > 0) { return $f->getReturn(); } $f->resume(); } }", 0},
		{"foreach repeats later resume", "<?php function a($xs) { " + alloc + "foreach ($xs as $i) { if ($i > 0) { return $f->getReturn(); } $f->resume(); } }", 0},
		{"property receiver", "<?php function a($o) { $o->f = new Fiber(fn() => Fiber::suspend()); $o->f->start(); return $o->f->getReturn(); }", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte(tc.src), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}

type reviewProbe struct {
	id    string
	kinds []syntax.NodeKind
	check func(*analysis.Context, syntax.Node)
}

func (p reviewProbe) ID() string { return p.id }

func (p reviewProbe) Semantic() {}

func (p reviewProbe) Flow() {}

func (p reviewProbe) Kinds() []syntax.NodeKind { return p.kinds }

func (p reviewProbe) Check(ctx *analysis.Context, n syntax.Node) { p.check(ctx, n) }
