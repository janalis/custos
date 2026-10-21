package sessionlockheldduringblockingcall

import (
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewSessionLockHeldDuringBlockingCall(t *testing.T) {
	const files = "ini_set('session.save_handler', 'files');"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"plain start", files + "session_start(); sleep(5);", 1},
		{"inside function", "function f(){" + files + "session_start(); sleep(5); usleep(microseconds: 10);}", 2},
		{"network transfer", "function f($ch){" + files + "session_start(); curl_exec($ch);}", 1},
		{"closed before blocking", files + "session_start(); session_write_close(); sleep(5);", 0},
		{"no session", files + "sleep(5);", 0},
		{"no handler setting", "session_start(); sleep(5);", 0},
		{"user function shadowing sleep", "namespace App; function sleep($s){} \\ini_set('session.save_handler', 'files'); \\session_start(); sleep(5);", 0},

		{"read and close", files + "session_start(['read_and_close' => true]); sleep(5);", 0},
		{"named read and close", files + "session_start(options: ['read_and_close' => 1]); sleep(5);", 0},
		{"read and close through variable", files + "$o = ['read_and_close' => true]; session_start($o); sleep(5);", 0},
		{"read and close disabled", files + "session_start(['read_and_close' => false]); sleep(5);", 1},
		{"dynamic read and close", "function f($c){" + files + "session_start(['read_and_close' => $c]); sleep(5);}", 0},
		{"unrelated option", files + "session_start(['cookie_lifetime' => 60]); sleep(5);", 1},
		{"unrelated dynamic option", "function f($n){" + files + "session_start(['name' => $n]); sleep(5);}", 1},
		{"empty options", files + "session_start([]); sleep(5);", 1},
		{"dynamic options", "function f($o){" + files + "session_start($o); sleep(5);}", 0},
		{"spread options", "function f($o){" + files + "session_start(...$o); sleep(5);}", 0},
		{"unpacked option array", "function f($o){" + files + "session_start([...$o]); sleep(5);}", 0},
		{"dynamic option key", "function f($k){" + files + "session_start([$k => true]); sleep(5);}", 0},
		{"positional option entry", files + "session_start([true]); sleep(5);", 0},

		{"handler replaced", files + "ini_set('session.save_handler', 'redis'); session_start(); sleep(5);", 0},
		{"handler restored to files", "ini_set('session.save_handler', 'redis');" + files + "session_start(); sleep(5);", 1},
		{"alias sets files", "ini_alter('session.save_handler', 'files'); session_start(); sleep(5);", 1},
		{"dynamic handler value", "function f($h){" + files + "ini_set('session.save_handler', $h); session_start(); sleep(5);}", 0},
		{"dynamic option name", "function f($k){" + files + "ini_set($k, 'redis'); session_start(); sleep(5);}", 0},
		{"other option", files + "ini_set('session.gc_maxlifetime', '60'); session_start(); sleep(5);", 1},
		{"setting restored", files + "ini_restore('session.save_handler'); session_start(); sleep(5);", 0},
		{"other setting restored", files + "ini_restore('memory_limit'); session_start(); sleep(5);", 1},
		{"custom handler", "function f($h){" + files + "session_set_save_handler($h, true); session_start(); sleep(5);}", 0},
		{"module selected", files + "session_module_name('redis'); session_start(); sleep(5);", 0},
		{"module read", files + "$m = session_module_name(); session_start(); sleep(5);", 1},
		{"conditional files handler", "function f($c){if($c){" + files + "} session_start(); sleep(5);}", 0},
		{"conditional replacement", "function f($c){" + files + "if($c){ini_set('session.save_handler', 'redis');} session_start(); sleep(5);}", 0},
		{"handler written after start", "session_start();" + files + "sleep(5);", 0},
		{"replacement attempted after start", files + "session_start(); ini_set('session.save_handler', 'redis'); sleep(5);", 1},
		{"method call before start", "function f($o){" + files + "$o->prepare(); session_start(); sleep(5);}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := analyzeReview(t, tc.src)
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}

func analyzeReview(t *testing.T, src string) []diagnostic.Finding {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"SessionLockHeldDuringBlockingCall"}})
	if err != nil {
		t.Fatal(err)
	}
	return e.Analyze(syntax.Parse("review.php", []byte("<?php "+src), syntax.Options{}))
}
