package transliterationfailureconsumed

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
	const produce = "<?php function f($t, $i, $o) { $s = transliterator_transliterate($t, $i); "
	for _, tc := range []struct {
		name, body string
		findings   int
	}{
		{"unguarded", "return strlen($s);", 1},
		{"direct", "return strlen(transliterator_transliterate($t, $i));", 1},
		{"if body", "if ($s !== false) { return strlen($s); }", 0},
		{"if body unrelated", "if ($o) { return strlen($s); }", 1},
		{"if condition", "if (strlen($s) > 1) { return 1; }", 1},
		{"if condition guarded", "if ($s !== false && strlen($s) > 1) { return 1; }", 0},
		{"else", "if ($s === false) { return 0; } else { return strlen($s); }", 0},
		{"else unrelated", "if ($o) { return 0; } else { return strlen($s); }", 1},
		{"elseif condition", "if ($s === false) { return 0; } elseif (strlen($s) > 1) { return 1; }", 0},
		{"elseif body", "if ($o) { return 0; } elseif ($s !== false) { return strlen($s); }", 0},
		{"elseif body unrelated", "if ($o) { return 0; } elseif ($o > 1) { return strlen($s); }", 1},
		{"after failed elseif", "if ($o) { return 0; } elseif ($s === false) { return 1; } elseif ($o > 1) { return strlen($s); }", 0},
		{"else after failed elseif", "if ($o) { return 0; } elseif ($s === false) { return 1; } else { return strlen($s); }", 0},
		{"before later elseif", "if ($o) { return 0; } elseif ($o > 1) { return strlen($s); } elseif ($s === false) { return 1; }", 1},
		{"else after type check", "if (!is_string($s)) { return 0; } else { return strlen($s); }", 0},
		{"elseif after type check", "if (!is_string($s)) { return 0; } elseif (strlen($s) > 1) { return 1; }", 0},
		{"elseif type check", "if ($o) { return 0; } elseif (is_string($s)) { return strlen($s); }", 0},
		{"else after elseif type check", "if ($o) { return 0; } elseif (!is_string($s)) { return 1; } else { return strlen($s); }", 0},
		{"elseif before type check", "if ($o) { return 0; } elseif ($o > 1) { return strlen($s); } elseif (!is_string($s)) { return 1; }", 1},
		{"ternary else", "return $s === false ? 0 : strlen($s);", 0},
		{"ternary then", "return $s !== false ? strlen($s) : 0;", 0},
		{"ternary yoda", "return false === $s ? 0 : strlen($s);", 0},
		{"ternary wrong branch", "return $s === false ? strlen($s) : 0;", 1},
		{"ternary wrong else", "return $s !== false ? 0 : strlen($s);", 1},
		{"ternary condition", "return strlen($s) ? 1 : 0;", 1},
		{"short ternary", "return $o ?: strlen($s);", 1},
		{"truthy ternary", "return $s ? strlen($s) : 0;", 0},
		{"loose ternary", "return $s != false ? strlen($s) : 0;", 0},
		{"null ternary", "return $s !== null ? strlen($s) : 0;", 1},
		{"other variable", "return $o !== false ? strlen($s) : 0;", 1},
		{"and", "return $s !== false && strlen($s) > 3;", 0},
		{"keyword and", "return ($s !== false and strlen($s) > 3);", 0},
		{"or", "return $s === false || strlen($s) > 3;", 0},
		{"keyword or", "return ($s === false or strlen($s) > 3);", 0},
		{"and wrong sense", "return $s === false && strlen($s) > 3;", 1},
		{"or wrong sense", "return $s !== false || strlen($s) > 3;", 1},
		{"left operand", "return strlen($s) > 3 && $s !== false;", 1},
		{"coalesce", "return $o ?? strlen($s);", 1},
		{"nested and", "return $o && $s !== false && strlen($s) > 3;", 0},
		{"conjunction then", "return ($o && $s !== false) ? strlen($s) : 0;", 0},
		{"conjunction else", "return ($o && $s === false) ? 0 : strlen($s);", 1},
		{"disjunction else", "return ($o || $s === false) ? 0 : strlen($s);", 0},
		{"disjunction then", "return ($o || $s !== false) ? strlen($s) : 0;", 1},
		{"both disjuncts", "return ($s !== false || is_string($s)) ? strlen($s) : 0;", 0},
		{"both conjuncts fail", "return ($s === false && !is_string($s)) ? 0 : strlen($s);", 0},
		// The flow does not combine two operands that both exclude false: the
		// rule's own walk decides the rows below.
		{"conjunction inside both disjuncts", "return (($o && $s !== false) || is_string($s)) ? strlen($s) : 0;", 0},
		{"conjunction inside one disjunct", "return (($o && $s !== false) || is_string($o)) ? strlen($s) : 0;", 1},
		{"disjunction inside both failed conjuncts", "return (($o || $s === false) && !is_string($s)) ? 0 : strlen($s);", 0},
		{"disjunction inside one failed conjunct", "return (($o || $s === false) && !is_string($o)) ? 0 : strlen($s);", 1},
		{"match arm of both disjuncts", "return match (true) { $s !== false || is_string($s) => strlen($s), default => 0 };", 0},
		{"match arm of one disjunct", "return match (true) { $s !== false || is_string($o) => strlen($s), default => 0 };", 1},
		{"match default after both conjuncts fail", "return match (true) { $s === false && !is_string($s) => 0, default => strlen($s) };", 0},
		{"match default after one conjunct fails", "return match (true) { $s === false && !is_string($o) => 0, default => strlen($s) };", 1},
		{"both disjuncts branch", "if ($s !== false || is_string($s)) { return strlen($s); }", 0},
		{"one disjunct branch", "if ($s !== false || is_string($o)) { return strlen($s); }", 1},
		{"both conjuncts fail else", "if ($s === false && !is_string($s)) { return 0; } else { return strlen($s); }", 0},
		{"one conjunct fails else", "if ($s === false && !is_string($o)) { return 0; } else { return strlen($s); }", 1},
		{"else after both conjuncts fail in elseif", "if ($o) { return 0; } elseif ($s === false && !is_string($s)) { return 1; } else { return strlen($s); }", 0},
		{"else after one conjunct fails in elseif", "if ($o) { return 0; } elseif ($s === false && !is_string($o)) { return 1; } else { return strlen($s); }", 1},
		{"both disjuncts elseif", "if ($o) { return 0; } elseif ($s !== false || is_string($s)) { return strlen($s); }", 0},
		{"both disjuncts and", "return ($s !== false || is_string($s)) && strlen($s) > 3;", 0},
		{"both conjuncts fail or", "return ($s === false && !is_string($s)) || strlen($s) > 3;", 0},
		{"preceding both conjuncts", "if ($s === false && !is_string($s)) { return 0; } return strlen($s);", 0},
		{"preceding one conjunct", "if ($s === false && !is_string($o)) { return 0; } return strlen($s);", 1},
		{"match conjunction", "return match (true) { $o && $s !== false => strlen($s), default => 0 };", 0},
		{"match failed disjunction", "return match (true) { $o || $s === false => 0, default => strlen($s) };", 0},
		{"match disjunction", "return match (true) { $o || $s !== false => strlen($s), default => 0 };", 1},
		{"match failed conjunction", "return match (true) { $o && $s === false => 0, default => strlen($s) };", 1},
		{"negated", "return !($s !== false) ? 0 : strlen($s);", 0},
		{"negative number", "return -$o ? 0 : strlen($s);", 1},
		{"is_string", "return is_string($s) ? strlen($s) : 0;", 0},
		{"not is_string", "if (!is_string($s)) { return 0; } return strlen($s);", 0},
		{"is_string wrong sense", "return is_string($s) ? 0 : strlen($s);", 1},
		{"is_string other", "return is_string($o) ? strlen($s) : 0;", 1},
		{"is_bool", "return is_bool($s) ? 0 : strlen($s);", 0},
		{"match default", "return match (true) { $s === false => 0, default => strlen($s) };", 0},
		{"match own condition", "return match (true) { $s !== false => strlen($s), default => 0 };", 0},
		{"match later condition", "return match (true) { $s === false => 0, strlen($s) > 3 => 1, default => 2 };", 0},
		{"match first condition", "return match (true) { strlen($s) > 3 => 1, default => 2 };", 1},
		{"match two conditions", "return match (true) { $s !== false, $o => strlen($s), default => 0 };", 1},
		{"match unrelated arm", "return match (true) { $o => 0, default => strlen($s) };", 1},
		{"match later guard", "return match (true) { $o => strlen($s), $s === false => 0, default => 1 };", 1},
		{"match subject", "return match ($o) { $s === false => 0, default => strlen($s) };", 1},
		{"match on consumer", "return match (strlen($s)) { 0 => 0, default => 1 };", 1},
		{"preceding guard", "if ($s === false) { return 0; } return strlen($s);", 0},
		{"preceding throw", "if (false === $s) { throw new E(); } return strtoupper($s);", 0},
		{"preceding mixed guard", "if ($s === false || $s === '') { return 0; } return strlen($s);", 0},
		{"preceding fallthrough", "if ($s === false) { log(); } return strlen($s);", 1},
		{"preceding with else", "if ($s === false) { return 0; } else { log(); } return strlen($s);", 0},
		{"preceding with elseif", "if ($s === false) { return 0; } elseif ($o) { log(); } return strlen($s);", 0},
		{"preceding truthiness", "if (!$s) { return 0; } return strlen($s);", 0},
		{"rebound after guard", "if ($s === false) { return 0; } $s = transliterator_transliterate($t, $o); return strlen($s);", 1},
		{"rebound in branch", "if ($s !== false) { $s = transliterator_transliterate($t, $o); return strlen($s); }", 1},
		{"rebound in ternary", "return $s !== false ? [$s = transliterator_transliterate($t, $o), strlen($s)] : 0;", 1},
		{"closure", "return function () use ($s) { return strlen($s); };", 0},
	} {
		t.Run(tc.name, func(t *testing.T) { probe(t, e, produce+tc.body+" }", tc.findings) })
	}
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"assigned in guard", "<?php function f($t, $i) { if (($s = transliterator_transliterate($t, $i)) === false) { return 0; } return strlen($s); }", 0},
		{"assigned yoda", "<?php function f($t, $i) { if (false === ($s = transliterator_transliterate($t, $i))) { return 0; } return strlen($s); }", 0},
		{"assigned in branch", "<?php function f($t, $i) { if (($s = transliterator_transliterate($t, $i)) !== false) { return strlen($s); } }", 0},
		{"assigned method", "<?php function f(Transliterator $t, $i) { return ($s = $t->transliterate($i)) === false ? '' : strtoupper($s); }", 0},
		{"assigned truthiness", "<?php function f($t, $i) { if (!($s = transliterator_transliterate($t, $i))) { return 0; } return strlen($s); }", 0},
		{"assigned other", "<?php function f($t, $i) { $s = transliterator_transliterate($t, $i); if (($o = transliterator_transliterate($t, $i)) === false) { return 0; } return strlen($s); }", 1},
		{"assigned then rebound", "<?php function f($t, $i) { if (($s = transliterator_transliterate($t, $i)) === false) { return 0; } $s = transliterator_transliterate($t, $i); return strlen($s); }", 1},
		{"compound assignment", "<?php function f($t, $i) { $s = transliterator_transliterate($t, $i); if (($s .= '') === false) { return 0; } return strlen($s); }", 0},
		{"reference assignment", "<?php function f($t, $i, $o) { $s = transliterator_transliterate($t, $i); if (($s = &$o) === false) { return 0; } return strlen($s); }", 0},
		{"shadowed is_string", "<?php namespace App; function is_string($v) { return true; } function f($t, $i) { $s = \\transliterator_transliterate($t, $i); return is_string($s) ? \\strlen($s) : 0; }", 1},
		{"file scope", "<?php $s = transliterator_transliterate($t, $i); echo $s === false ? 0 : strlen($s);", 0},
	} {
		t.Run(tc.name, func(t *testing.T) { probe(t, e, tc.source, tc.findings) })
	}
}

func probe(t *testing.T, e *analysis.Engine, source string, findings int) {
	t.Helper()
	f := syntax.ParseBest("review.php", []byte(source), syntax.Options{Version: phpversion.PHP85})
	if len(f.Errors) != 0 {
		t.Fatalf("invalid probe: %v", f.Errors)
	}
	if got := e.Analyze(f); len(got) != findings {
		t.Fatalf("findings: got %v, want %d", got, findings)
	}
}
