package config

import (
	"encoding/json"
	"testing"
)

func TestFlowAssertionConfiguration(t *testing.T) {
	for _, tc := range []struct {
		json string
		want bool
	}{
		{`{}`, false}, {`{"flow":{}}`, false}, {`{"flow":{"assertions":false}}`, false}, {`{"flow":{"assertions":true}}`, true},
	} {
		var file File
		if err := json.Unmarshal([]byte(tc.json), &file); err != nil {
			t.Fatal(err)
		}
		cfg, err := Resolve(t.TempDir(), file)
		if err != nil || cfg.Flow.Assertions != tc.want || cfg.Analysis().Flow.Assertions != tc.want {
			t.Fatalf("config=%+v err=%v", cfg, err)
		}
	}
	var file File
	if err := json.Unmarshal([]byte(`{"flow":{"assertions":"true"}}`), &file); err == nil {
		t.Fatal("nonboolean option accepted")
	}
}
