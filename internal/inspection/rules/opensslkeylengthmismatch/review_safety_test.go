package opensslkeylengthmismatch

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	const use = " return openssl_encrypt($d, 'aes-256-cbc', $key, 0, $iv); }"
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"local literal", "function f($d, $iv) { $key = 'unset'; load();" + use, 1},
		{"parenthesised key", "function f($d, $iv) { $key = 'unset'; return openssl_encrypt($d, 'aes-256-cbc', ($key), 0, $iv); }", 1},
		{"value capture", "function f($d, $iv) { $key = 'unset'; $g = function () use ($key) { return load(); }; $g();" + use, 1},
		{"arrow capture", "function f($d, $iv) { $key = 'unset'; $g = fn () => $key = load(); $g();" + use, 1},
		{"other reference capture", "function f($d, $iv) { $key = 'unset'; $g = function () use (&$other) { $other = load(); }; $g();" + use, 1},
		{"reference capture", "function f($d, $iv) { $key = 'unset'; $g = function () use (&$key) { $key = load(); }; $g();" + use, 0},
		{"reference capture after another", "function f($d, $iv) { $key = 'unset'; $g = function () use (&$key) {}; $h = function () use (&$key) {}; $g();" + use, 0},
		{"reference item", "function f($d, $iv) { $key = 'unset'; $refs = [&$key]; fill($refs);" + use, 0},
		{"other reference item", "function f($d, $iv) { $key = 'unset'; $refs = [&$other, $key]; fill($refs);" + use, 1},
		{"extract", "function f($d, $iv) { $key = 'unset'; extract(load());" + use, 0},
		{"include", "function f($d, $iv) { $key = 'unset'; include 'key.php';" + use, 0},
		{"eval", "function f($d, $iv) { $key = 'unset'; eval($code);" + use, 0},
		{"global then call", "function f($d, $iv) { global $key; $key = 'unset'; load();" + use, 0},
		{"global then method", "function f($d, $iv, $o) { global $key; $key = 'unset'; $o->load();" + use, 0},
		{"global then static call", "function f($d, $iv) { global $key; $key = 'unset'; Keys::load();" + use, 0},
		{"global then new", "function f($d, $iv) { global $key; $key = 'unset'; new Loader();" + use, 0},
		{"global then superglobal", "function f($d, $iv, $v) { global $key; $key = 'unset'; $GLOBALS['key'] = $v;" + use, 0},
		{"global call in arguments", "function f($d, $iv) { global $key; $key = 'unset'; return openssl_encrypt(load(), 'aes-256-cbc', $key, 0, $iv); }", 0},
		{"global rebinding", "function f($d, $iv) { $key = 'unset'; global $key;" + use, 0},
		{"global without call", "function f($d, $iv) { global $other, $key; prepare(); $key = str_repeat('k', 5);" + use, 1},
		{"other global", "function f($d, $iv) { global $other; $key = 'unset'; load();" + use, 1},
		{"reference parameter", "function f($d, $iv, &$key) { $key = 'unset'; load();" + use, 0},
		{"reference parameter without call", "function f($d, $iv, &$key) { $key = 'unset';" + use, 1},
		{"captured scope", "$g = function ($d, $iv) use (&$key) { $key = 'unset'; load();" + use + ";", 0},
		{"captured scope without call", "$g = function ($d, $iv) use (&$key, $o) { $key = 'unset';" + use + ";", 1},
		{"file scope call", "$key = 'unset'; load(); openssl_encrypt($d, 'aes-256-cbc', $key, 0, $iv);", 0},
		{"file scope builtin", "$key = 'unset'; $iv = random_bytes(16); echo base64_encode(openssl_encrypt($d, 'aes-256-cbc', $key, 0, substr($iv, 0, 16)));", 1},
		{"file scope callback builtin", "$key = 'unset'; usort($list, 'load'); openssl_encrypt($d, 'aes-256-cbc', $key, 0, $iv);", 0},
		{"global then builtin", "function f($d, $iv) { global $key; $key = 'unset'; trim($d);" + use, 1},
		{"file scope direct", "prepare(); $key = 'unset'; openssl_encrypt($d, 'aes-256-cbc', $key, 0, $iv);", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte("<?php "+tc.source), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
