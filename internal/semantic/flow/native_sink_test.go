package flow

import "testing"

func TestNativeSinkSelection(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		contexts     []Context
	}{
		{"ordinary builtin", `strlen('text');`, nil},
		{"user function", `function custom($x){} custom('text');`, nil},
		{"named shell command", `system(command: 'text');`, []Context{Shell}},
		{"reference output", `$output=[]; exec('text', $output);`, []Context{Shell}},
		{"argument vector", `proc_open(['text'], [], $pipes);`, nil},
		{"location header", `header('Location: '.$_GET['url']);`, []Context{Header, URL}},
		{"curl URL", `curl_setopt(null, \CURLOPT_URL, 'text');`, []Context{URL}},
		{"named curl URL", `curl_setopt(value: 'text', option: CURLOPT_URL, handle: null);`, []Context{URL}},
		{"other curl option", `curl_setopt(null, CURLOPT_TIMEOUT, 1);`, nil},
		{"dynamic curl option", `curl_setopt(null, $_GET['option'], 'text');`, nil},
		{"incomplete curl arguments", `curl_setopt(null);`, nil},
		{"missing PDO query argument", `$db = new PDO('sqlite::memory:'); $db->query();`, nil},
		{"PDO query", `$db = new PDO('sqlite::memory:'); $db->query('select 1');`, []Context{SQL}},
		{"ordinary native method", `$date = new DateTime(); $date->format('Y');`, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, _ := environment(t, tc.source, nil)
			var contexts []Context
			for _, call := range e.Calls(nil) {
				for _, sink := range e.nativeSinks(call) {
					contexts = append(contexts, sink.Context)
				}
			}
			if len(contexts) != len(tc.contexts) {
				t.Fatalf("contexts = %v, want %v", contexts, tc.contexts)
			}
			for i, context := range contexts {
				if context != tc.contexts[i] {
					t.Fatalf("contexts = %v, want %v", contexts, tc.contexts)
				}
			}
		})
	}
}
