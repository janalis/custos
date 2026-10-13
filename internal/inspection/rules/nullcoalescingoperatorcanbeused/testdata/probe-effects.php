<?php
// The probe runs twice here and once with `??`: reported, never rewritten.
function nextId() { static $n = 0; return ++$n; }
$a = <weak_warning descr="Simplify to 'nextId() ?? 'none'' using the null coalescing operator.">nextId() !== null ? nextId() : 'none'</weak_warning>;
$b = <weak_warning descr="Simplify to '$rows[nextId()] ?? 'none'' using the null coalescing operator.">isset($rows[nextId()]) ? $rows[nextId()] : 'none'</weak_warning>;
$c = <weak_warning descr="Simplify to '$rows[++$i] ?? null' using the null coalescing operator.">array_key_exists(++$i, $rows) ? $rows[++$i] : null</weak_warning>;
function firstId() {
    <weak_warning descr="Simplify to 'return nextId() ?? 'none'' using the null coalescing operator.">if</weak_warning> (nextId() !== null) {
        return nextId();
    }
    return 'none';
}
// A fallback with effects is evaluated at most once in both forms.
$d = <weak_warning descr="Simplify to '$rows[$k] ?? nextId()' using the null coalescing operator.">isset($rows[$k]) ? $rows[$k] : nextId()</weak_warning>;
