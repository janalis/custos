package preloadingusagecorrectness

import (
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source string
		findings     int
		fixed        string
	}{
		// The included file must run: later code reads what it defines.
		{"variable read", "<?php require 'p.php'; foreach ($paths as $p) { echo $p; }", 0, ""},
		{"variable merged", "<?php require 'p.php'; $paths = array_merge($paths, ['x']);", 0, ""},
		{"variable appended", "<?php require 'p.php'; $paths .= 'x';", 0, ""},
		{"variable counted", "<?php require 'p.php'; $n++;", 0, ""},
		{"element written", "<?php require 'c.php'; $config['debug'] = true;", 0, ""},
		{"captured by closure", "<?php require 'p.php'; $f = function () use ($paths) { return $paths; };", 0, ""},
		{"read in arrow function", "<?php require 'p.php'; $f = fn () => $paths;", 0, ""},
		{"read in function scope", "<?php function warm($dir) { require $dir; return $loaded; }", 0, ""},
		{"read after enclosing block", "<?php if (PHP_VERSION_ID > 80000) { require 'p.php'; } foreach ($paths as $p) { echo $p; }", 0, ""},
		{"read after unbraced if", "<?php if (PHP_VERSION_ID > 80000) require 'p.php'; echo $paths;", 0, ""},
		{"read after block in function", "<?php function warm($on) { if ($on) { require 'p.php'; } return $loaded; }", 0, ""},
		{"read after block in namespace", "<?php namespace A { if (PHP_VERSION_ID > 80000) { require 'p.php'; } echo $paths; }", 0, ""},
		{"constant read", "<?php require 'c.php'; echo APP_ROOT, PHP_EOL;", 0, ""},
		{"constant before builtin", "<?php require 'c.php'; echo [APP_ROOT], strlen('x');", 0, ""},
		{"namespaced constant", "<?php namespace App; require 'c.php'; echo ROOT;", 0, ""},
		{"imported constant", "<?php use const Lib\\ROOT; require 'c.php'; echo ROOT;", 0, ""},
		{"defined probe", "<?php require 'c.php'; echo defined('APP_ROOT') ? 1 : 0;", 0, ""},
		{"constant lookup", "<?php require 'c.php'; echo constant('\\\\APP_ROOT');", 0, ""},
		{"defined later", "<?php require 'c.php'; echo APP_ROOT; define('APP_ROOT', 1);", 0, ""},
		{"declared in another namespace", "<?php namespace A { const ROOT = 1; } namespace B { require 'c.php'; echo \\B\\ROOT; }", 0, ""},
		// The script provides everything the later code reads.
		{
			"own variable", "<?php $paths = ['a']; require 'p.php'; foreach ($paths as $p) { echo $p; }", 1,
			"<?php $paths = ['a']; opcache_compile_file('p.php'); foreach ($paths as $p) { echo $p; }",
		},
		{
			"assigned after", "<?php require 'p.php'; $n = 1; $n++; unset($n);", 1,
			"<?php opcache_compile_file('p.php'); $n = 1; $n++; unset($n);",
		},
		{
			"parameter", "<?php function warm($dir) { require $dir; return [$dir, $this ?? null, $argv]; }", 1,
			"<?php function warm($dir) { opcache_compile_file($dir); return [$dir, $this ?? null, $argv]; }",
		},
		{
			"closure locals", "<?php require 'p.php'; spl_autoload_register(function ($class) { $file = $class; });", 1,
			"<?php opcache_compile_file('p.php'); spl_autoload_register(function ($class) { $file = $class; });",
		},
		{
			"own constants", "<?php namespace App; const ROOT = 1; define('MODE', 2); require 'c.php'; echo ROOT, MODE, null, defined('MODE'), define(...);", 1,
			"<?php namespace App; const ROOT = 1; define('MODE', 2); opcache_compile_file('c.php'); echo ROOT, MODE, null, defined('MODE'), define(...);",
		},
		{
			"dynamic constant name", "<?php $n = 'X'; require 'c.php'; echo defined($n), defined(('M') . 'N'), defined(), E_ALL;", 1,
			"<?php $n = 'X'; opcache_compile_file('c.php'); echo defined($n), defined(('M') . 'N'), defined(), E_ALL;",
		},
		// Comments in the replaced range are kept.
		{
			"comment after keyword", "<?php require /* first */ __DIR__ . '/a.php';", 1,
			"<?php opcache_compile_file(/* first */ __DIR__ . '/a.php');",
		},
		{
			"comment glued to keyword", "<?php require/* first */'a.php';", 1,
			"<?php opcache_compile_file(/* first */'a.php');",
		},
		{
			"line comment in parentheses", "<?php require( // generated\n    'a.php');", 1,
			"<?php opcache_compile_file( // generated\n    'a.php');",
		},
		{
			"comment before closing parenthesis", "<?php include ('a.php' /* optional */);", 1,
			"<?php opcache_compile_file('a.php' /* optional */);",
		},
		{
			"comment in partly parenthesised path", "<?php include (__DIR__ /* here */) . '/a.php';", 1,
			"<?php opcache_compile_file((__DIR__ /* here */) . '/a.php');",
		},
		{
			"comment inside the argument", "<?php include (__DIR__ . /* here */ '/a.php');", 1,
			"<?php opcache_compile_file(__DIR__ . /* here */ '/a.php');",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(tc.source)
			opt := syntax.Options{Version: phpversion.PHP85}
			f := syntax.ParseBest("app/preload.php", src, opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			got := e.Analyze(f)
			if len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
			res := fixing.FixSource(e, "app/preload.php", src, fixing.Options{Parse: opt, SinglePass: true})
			want := tc.source
			if tc.fixed != "" {
				want = tc.fixed
			}
			if string(res.Source) != want {
				t.Fatalf("fixed source: got %s, want %s", res.Source, want)
			}
			if len(syntax.ParseBest("app/preload.php", res.Source, opt).Errors) != 0 {
				t.Fatal("fix introduced syntax errors")
			}
		})
	}
}
