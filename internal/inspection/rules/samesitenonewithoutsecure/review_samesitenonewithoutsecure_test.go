package samesitenonewithoutsecure

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewSameSiteNoneWithoutSecure(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"secure false", "setcookie('a','x',['samesite'=>'None','secure'=>false]);", 1},
		{"secure absent", "setcookie('a','x',['samesite'=>'None']);", 1},
		{"capitalised secure true", "setcookie('a','x',['samesite'=>'None','Secure'=>true]);", 0},
		{"upper secure true", "setcookie('a','x',['samesite'=>'None','SECURE'=>true,'path'=>'/']);", 0},
		{"capitalised secure false", "setcookie('a','x',['samesite'=>'None','Secure'=>false]);", 1},
		{"mixed case samesite", "setcookie('a','x',['SameSite'=>'None']);", 1},
		{"mixed case both", "setcookie('a','x',['SAMESITE'=>'none','sEcUrE'=>0]);", 1},
		{"mixed case samesite lax", "setcookie('a','x',['SameSite'=>'Lax']);", 0},
		{"secure spelled twice", "setcookie('a','x',['samesite'=>'None','secure'=>false,'Secure'=>true]);", 0},
		{"samesite spelled twice", "setcookie('a','x',['samesite'=>'None','SameSite'=>'Lax']);", 0},
		{"unknown secure", "setcookie('a','x',['samesite'=>'None','Secure'=>$s]);", 0},
		{"unknown options", "setcookie('a','x',$options);", 0},
		{"legacy expires argument", "setcookie('a','x',0);", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probe := New()
			e, err := analysis.NewEngine([]analysis.Rule{probe}, analysis.Config{Only: []string{"SameSiteNoneWithoutSecure"}})
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
