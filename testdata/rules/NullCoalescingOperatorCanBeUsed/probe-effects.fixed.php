<?php
// The probe runs twice here and once with `??`: reported, never rewritten.
function nextId() { static $n = 0; return ++$n; }
$a = nextId() !== null ? nextId() : 'none';
$b = isset($rows[nextId()]) ? $rows[nextId()] : 'none';
$c = array_key_exists(++$i, $rows) ? $rows[++$i] : null;
function firstId() {
    if (nextId() !== null) {
        return nextId();
    }
    return 'none';
}
// A fallback with effects is evaluated at most once in both forms.
$d = $rows[$k] ?? nextId();
