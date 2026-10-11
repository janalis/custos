<?php

function wrongBranch(NumberFormatter $f, $n): int
{
    $s = $f->format($n);
    if ($s === false) {
        return strlen(<warning descr="Check formatting success before consuming the formatted text.">$s</warning>);
    }
    return 0;
}

function unrelatedGuard(NumberFormatter $f, $n, $other): int
{
    $s = $f->format($n);
    if ($other === false || !is_string($other) || $n) {
        return 0;
    }
    return strlen(<warning descr="Check formatting success before consuming the formatted text.">$s</warning>);
}

function guardDoesNotLeave(NumberFormatter $f, $n): int
{
    $s = $f->format($n);
    if ($s === false) {
        $n++;
    }
    return strlen(<warning descr="Check formatting success before consuming the formatted text.">$s</warning>);
}

function staleGuard(NumberFormatter $f, $n): int
{
    $s = $f->format($n);
    if ($s === false) {
        return 0;
    }
    $s = $f->format($n + 1);
    return strlen(<warning descr="Check formatting success before consuming the formatted text.">$s</warning>);
}

function weakTests(NumberFormatter $f, $n): int
{
    $s = $f->format($n);
    $a = isset($s) ? strlen(<warning descr="Check formatting success before consuming the formatted text.">$s</warning>) : 0;
    $b = $s !== null && strlen(<warning descr="Check formatting success before consuming the formatted text.">$s</warning>) > 1;
    $c = $s === false && strlen(<warning descr="Check formatting success before consuming the formatted text.">$s</warning>) > 1;
    $d = $s ?: strlen(<warning descr="Check formatting success before consuming the formatted text.">$s</warning>);
    $e = ($s || $n) ? strlen(<warning descr="Check formatting success before consuming the formatted text.">$s</warning>) : 0;
    $g = -$s ? strlen(<warning descr="Check formatting success before consuming the formatted text.">$s</warning>) : 0;
    $h = is_numeric($s) ? strlen(<warning descr="Check formatting success before consuming the formatted text.">$s</warning>) : 0;
    $i = $s > false ? strlen(<warning descr="Check formatting success before consuming the formatted text.">$s</warning>) : 0;
    return $a + $b + $c + $d + $e + $g + $h + $i;
}

function conditionUse(NumberFormatter $f, $n): int
{
    $s = $f->format($n);
    if (strlen(<warning descr="Check formatting success before consuming the formatted text.">$s</warning>) > 3) {
        return 1;
    } elseif (strlen(<warning descr="Check formatting success before consuming the formatted text.">$s</warning>) > 2) {
        return 2;
    } elseif ($n) {
        return strlen(<warning descr="Check formatting success before consuming the formatted text.">$s</warning>);
    } else {
        return strlen(<warning descr="Check formatting success before consuming the formatted text.">$s</warning>) + 1;
    }
}

function sideEffectOnly(NumberFormatter $f, $n): int
{
    $s = $f->format($n);
    $s !== false or $n++;
    $s !== false + 1;
    is_string($s);
    return strlen(<warning descr="Check formatting success before consuming the formatted text.">$s</warning>);
}

function partialElse(NumberFormatter $f, $n): int
{
    $s = $f->format($n);
    if ($s) {
        $n++;
    } elseif ($n) {
        $n--;
    } else {
        return 0;
    }
    return strlen(<warning descr="Check formatting success before consuming the formatted text.">$s</warning>);
}

function otherAssignment(NumberFormatter $f, $n, $t): int
{
    $s = $f->format($n);
    if (($t = $f->format($n)) && ($s = 'x') && $n) {
        return 0;
    }
    return strlen($s);
}
