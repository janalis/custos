<?php
function build(array $parts, &$out) {
    $memo = null;
    $join = function ($sep, &$acc) use (&$memo, $parts) {
        if ($memo !== null) {
            return $memo = implode($sep, $parts);
        }
        if ($acc === []) {
            return $acc = $parts;
        }
        <weak_warning descr="The assigned variable is never used after returning; return the value directly.">return $sep = strtoupper($sep);</weak_warning>
    };
    if ($out) {
        return $out = $join;
    }
    <weak_warning descr="The assigned variable is never used after returning; return the value directly.">return $parts = $join;</weak_warning>
}

function clear(array &$rows) {
    if (!$rows) { return; }
    $rows = [];
    <weak_warning descr="Redundant 'return;' at the end of the body; remove it.">return;</weak_warning>
    // trailing comment
}

class K {
    public function m() {
        <weak_warning descr="The assigned variable is never used after returning; return the value directly.">return $x = [
            1, // one
            2,
        ];</weak_warning>
    }
}
