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
        return strtoupper($sep);
    };
    if ($out) {
        return $out = $join;
    }
    return $join;
}

function clear(array &$rows) {
    if (!$rows) { return; }
    $rows = [];
    return;
    // trailing comment
}

class K {
    public function m() {
        return [
            1, // one
            2,
        ];
    }
}
