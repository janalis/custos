package semanticquery

import "testing"

func TestReviewRedirectContinuesProtectedExecution(t *testing.T) {
	const (
		decl     = "/** @custos-protected */ function deleteRecord(){} "
		redirect = "if(!$ok){header('Location: /signin');}"
		open     = decl + "function f($ok,$other){" + redirect
	)
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"falls through", open + "deleteRecord();}", 1},
		{"compared with true", open + "if($ok===true){deleteRecord();}}", 0},
		{"early return with else", open + "if(!$ok){return;}else{$x=1;}deleteRecord();}", 0},
		{"while guard", open + "while($ok){deleteRecord();break;}}", 0},
		{"match arm", open + "return match(true){(bool)$other=>0,$ok=>deleteRecord(),default=>1};}", 0},
		{"jump label after guard", open + "if(!$ok){return;}back:deleteRecord();goto back;}", 1},
		{"guard on another value", open + "if($other===true){deleteRecord();}}", 1},
		{"falls through at top level", decl + redirect + "deleteRecord();", 1},
		{"top level guard", decl + redirect + "if(!$ok){exit;}deleteRecord();", 0},
		{"single statement branch", decl + "function f($ok){if(!$ok)header('Location: /signin');deleteRecord();}", 1},
		{"nested deeper than the redirect", open + "foreach([1] as $_){deleteRecord();}}", 1},
		{"unrelated guard", open + "if($other){deleteRecord();}}", 1},
		{"unrelated early return", open + "if(!$other){return;}deleteRecord();}", 1},
		{"same unauthorized branch again", open + "if(!$ok){deleteRecord();}}", 1},
		{"call in a condition", open + "if(deleteRecord()){echo 1;}}", 1},
		{"call left of conjunction", open + "deleteRecord()&&$ok;}", 1},
		{"call in ternary condition", open + "return deleteRecord()?1:2;}", 1},
		{"call in elseif condition", open + "if($other){}elseif(deleteRecord()){}}", 1},
		{"call in loop condition", open + "while(deleteRecord()){}}", 1},
		{"disjunction does not prove", open + "if($ok||$other){deleteRecord();}}", 1},
		{"negated conjunction guard", open + "if(!($ok&&$other)){return;}if($other){}else{deleteRecord();}}", 0},
		{"false conjunction else arm", open + "if($ok&&$other){}else{deleteRecord();}}", 1},
		{"bitwise operator ignored", open + "if($ok&$other){deleteRecord();}}", 1},
		{"negation of another operator", open + "if(-$ok){deleteRecord();}}", 1},
		{"variable variable predicate", decl + "function f($n){if(!$$n){header('Location: /signin');}deleteRecord();}", 0},
		{"non-variable predicate", decl + "function f($u){if(!$u->ok){header('Location: /signin');}deleteRecord();}", 0},
		{"positive predicate", decl + "function f($ok){if($ok){header('Location: /home');}deleteRecord();}", 0},
		{"not a location header", decl + "function f($ok){if(!$ok){header('X-Denied: 1');log_it();}deleteRecord();}", 0},
		{"header only inside a closure", decl + "function f($ok){if(!$ok){$later=function(){header('Location: /signin');};}deleteRecord();}", 0},
		{"header after a closure", decl + "function f($ok){if(!$ok){$later=function(){};header('Location: /signin');}deleteRecord();}", 1},
		{"call inside a later closure", open + "return function(){deleteRecord();};}", 0},
		{"call inside an arrow function", open + "return fn()=>deleteRecord();}", 0},
		{"redirect with else", decl + "function f($ok){if(!$ok){header('Location: /signin');}else{echo 1;}deleteRecord();}", 0},
		{"redirect after the call", decl + "function f($ok){deleteRecord();" + redirect + "}", 0},

		{"exclusive branch", open + "if($ok){deleteRecord();}}", 0},
		{"exclusive parenthesised conjunction", open + "if(($ok)&&$other){deleteRecord();}}", 0},
		{"exclusive word conjunction", open + "if($other and $ok){deleteRecord();}}", 0},
		{"else of unauthorized", open + "if(!$ok){echo 1;}else{deleteRecord();}}", 0},
		{"else of disjunction", open + "if($other||!$ok){echo 1;}else{deleteRecord();}}", 0},
		{"else of word disjunction", open + "if($other or !$ok){echo 1;}else{deleteRecord();}}", 0},
		{"elseif on predicate", open + "if($other){echo 1;}elseif($ok){deleteRecord();}}", 0},
		{"elseif after unauthorized", open + "if(!$ok){echo 1;}elseif($other){deleteRecord();}}", 0},
		{"second elseif after unauthorized elseif", open + "if($other){}elseif(!$ok){}elseif($other>1){deleteRecord();}}", 0},
		{"second elseif after unrelated elseif", open + "if($other){}elseif($other>2){}elseif($other>1){deleteRecord();}}", 1},
		{"else after unauthorized elseif", open + "if($other){}elseif(!$ok){}else{deleteRecord();}}", 0},
		{"while on predicate", open + "while($ok){deleteRecord();break;}}", 0},
		{"early return", open + "if(!$ok){return;}deleteRecord();}", 0},
		{"early throw in nested block", open + "foreach([1] as $_){if(!$ok){throw new Exception();}deleteRecord();}}", 0},
		{"early return on disjunction", open + "if($other||!$ok){return;}deleteRecord();}", 0},
		{"guard before the redirect only", decl + "function f($ok){if(!$ok&&rand()){return;}" + redirect + "deleteRecord();}", 1},
		{"ternary then", open + "return $ok?deleteRecord():null;}", 0},
		{"ternary else", open + "return !$ok?null:deleteRecord();}", 0},
		{"ternary else on predicate", open + "return $ok?null:deleteRecord();}", 1},
		{"short ternary", open + "return !$ok?:deleteRecord();}", 0},
		{"short ternary on predicate", open + "return $ok?:deleteRecord();}", 1},
		{"conjunction", open + "$ok&&deleteRecord();}", 0},
		{"word conjunction", open + "$ok and deleteRecord();}", 0},
		{"disjunction", open + "!$ok||deleteRecord();}", 0},
		{"word disjunction", open + "!$ok or deleteRecord();}", 0},
		{"disjunction on predicate", open + "$ok||deleteRecord();}", 1},
		{"other binary operator", open + "return $ok.deleteRecord();}", 1},

		{"predicate reassigned in branch", decl + "function f($ok){if(!$ok){header('Location: /signin');$ok=login();}deleteRecord();}", 0},
		{"predicate reassigned after branch", open + "$ok=login();deleteRecord();}", 0},
		{"predicate assigned before branch", decl + "function f(){$ok=check();" + redirect + "deleteRecord();}", 1},
		{"predicate assigned after call", open + "deleteRecord();$ok=true;}", 1},
		{"predicate only read between", open + "echo $ok;deleteRecord();}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := auditNative(t, "RedirectContinuesProtectedExecution", CheckRedirectContinuesProtectedExecution, tc.src)
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}
