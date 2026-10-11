package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewSessionCookieOptionsSetAfterStart(t *testing.T) {
	const set = "session_set_cookie_params(3600);"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"plain start", "session_start();" + set, 1},
		{"named argument", "session_start();session_set_cookie_params(lifetime_or_options: 3600);", 1},
		{"empty options", "session_start([]);" + set, 1},
		{"unrelated options", "session_start(['cookie_lifetime'=>10]);" + set, 1},
		{"read_and_close false", "session_start(['read_and_close'=>false]);" + set, 1},
		{"read_and_close zero", "session_start(options: ['read_and_close'=>0]);" + set, 1},
		{"read_and_close true", "session_start(['read_and_close'=>true]);" + set, 0},
		{"read_and_close one", "session_start(['read_and_close'=>1]);" + set, 0},
		{"read_and_close unknown", "session_start(['read_and_close'=>$flag]);" + set, 0},
		{"unknown options", "session_start($options);" + set, 0},
		{"spread options", "session_start(...$options);" + set, 0},
		{"real start after read_and_close", "session_start(['read_and_close'=>true]);session_start();" + set, 1},
		{"configured first", set + "session_start();", 0},
		{"closed again", "session_start();session_write_close();" + set, 0},
		{"failed start branch", "if(!session_start()){" + set + "}", 0},
		{"failed start else", "if(session_start()){}else{" + set + "}", 0},
		{"failed start compared", "if(session_start()===false){" + set + "}", 0},
		{"failed start not true", "if(true!=session_start()){" + set + "}", 0},
		{"failed start early exit", "function f(){if(session_start()){return;}" + set + "}", 0},
		{"failed start in disjunction else", "if(session_start()||$x){}else{" + set + "}", 0},
		{"failed start in conjunction", "if(!session_start()&&$x){" + set + "}", 0},
		{"successful start branch", "if(session_start()){" + set + "}", 1},
		{"successful start after guard", "function f(){if(!session_start()){return;}" + set + "}", 1},
		{"undecided disjunction", "if(!session_start()||$x){" + set + "}", 1},
		{"undecided conjunction else", "if(session_start()&&$x){}else{" + set + "}", 1},
		{"restart in failure branch", "if(!session_start()){session_start();" + set + "}", 1},
		{"other call condition", "session_start();if(!$x){" + set + "}", 1},
		{"other operator condition", "session_start();if(-$x){" + set + "}", 1},
		{"arithmetic condition", "session_start();if($x+1){" + set + "}", 1},
		{"variable condition", "session_start();if($x===$y){" + set + "}", 1},
		{"status not active", "session_start();if(session_status()!==PHP_SESSION_ACTIVE){" + set + "}", 0},
		{"status not active loose", "session_start();if(PHP_SESSION_ACTIVE!=session_status()){" + set + "}", 0},
		{"status none", "session_start();if(session_status()===PHP_SESSION_NONE){" + set + "}", 0},
		{"status active else", "session_start();if(session_status()===PHP_SESSION_ACTIVE){}else{" + set + "}", 0},
		{"status active", "session_start();if(session_status()===PHP_SESSION_ACTIVE){" + set + "}", 1},
		{"status not none", "session_start();if(session_status()!==PHP_SESSION_NONE){" + set + "}", 1},
		{"status unknown value", "session_start();if(session_status()===$state){" + set + "}", 1},
		{"other status function", "session_start();if(strlen($x)===PHP_SESSION_NONE){" + set + "}", 1},
		{"status tested before start", "if(session_status()!==PHP_SESSION_ACTIVE){session_start();" + set + "}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probe := nativeCheckProbe{"SessionCookieOptionsSetAfterStart", CheckSessionCookieOptionsSetAfterStart}
			e, err := analysis.NewEngine([]analysis.Rule{probe}, analysis.Config{Only: []string{"SessionCookieOptionsSetAfterStart"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("review.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}
