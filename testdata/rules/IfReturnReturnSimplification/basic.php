<?php
function isAdult(int $age): bool {
    <warning descr="Return the condition directly: 'return $age >= 18'.">if</warning> ($age >= 18) {
        return true;
    }
    return false;
}

function isBlank(string $s): bool {
    <warning descr="Return the condition directly: 'return $s !== '''.">if</warning> (($s === '')) { return FALSE; }
    // fallthrough
    return TRUE;
}

function hasBoth($m, $n) {
    <warning descr="Return the condition directly: 'return !($m && $n)'.">if</warning> ($m && $n) { return false; }
    else { return true; }
}

function notWidget($w) {
    <warning descr="Return the condition directly: 'return !($w instanceof Widget)'.">if</warning> ($w instanceof Widget) {
        return false;
    } else {
        return true;
    }
}

function guarded($q) {
    if ($q === null) { log_it(); return true; }
    if ($q > 9) { return true; }
    return false;
}

function guardedElse($q) {
    if ($q === null) { return true; }
    <warning descr="Return the condition directly: 'return $q < 3'.">if</warning> ($q < 3) { return true; }
    else { return false; }
}

function notBinary($q) {
    if (is_int($q)) { return true; }
    return false;
}

function unbraced($q) {
    if ($q > 1) return true;
    return false;
}

function sameValue($q) {
    if ($q > 1) { return false; }
    return false;
}

function hasElseif($q) {
    if ($q > 1) { return true; }
    elseif ($q < 0) { return true; }
    else { return false; }
}

function alt($q) {
    <warning descr="Return the condition directly: 'return $q <= 3'.">if</warning> ($q > 3):
        return false;
    endif;
    return true;
}

function bang($q) {
    <warning descr="Return the condition directly: 'return $q == 3'.">if</warning> ($q <> 3) { return \false; } else { return True; }
}
