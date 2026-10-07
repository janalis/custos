<?php
function isAdult(int $age): bool {
    return $age >= 18;
}

function isBlank(string $s): bool {
    return $s !== '';
}

function hasBoth($m, $n) {
    return !($m && $n);
}

function notWidget($w) {
    return !($w instanceof Widget);
}

function guarded($q) {
    if ($q === null) { log_it(); return true; }
    if ($q > 9) { return true; }
    return false;
}

function guardedElse($q) {
    if ($q === null) { return true; }
    return $q < 3;
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
    return $q <= 3;
}

function bang($q) {
    return $q == 3;
}
