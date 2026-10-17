<?php
function logSpec($spec, $extra = null) { echo $spec; }
function fillOther($label, &$out) { $out = 'P1D'; }
function collect(...$all) { return $all; }

// Reads that cannot rewrite the variable keep the report.
function byValue() {
    $s = <error descr="Malformed DateInterval specification.">'1W'</error>;
    logSpec($s);
    logSpec(extra: $s, spec: 'a');
    logSpec('a', 'b', $s);
    collect(1, 2, $s);
    $copy = $s;
    $list = [$s, 'k' => $s];
    $f = function () use ($s) { return $s; };
    $g = function ($o) { extract($o); };
    $other = 'P1D';
    fillOther($s, $other);
    return [new DateInterval($s), new SplObjectStorage($s)];
}

// An unresolved callee may replace the argument through a reference parameter.
function unresolvedWriter() {
    $s = '1W';
    unknownLogger($s);
    return new DateInterval($s);
}
function throughCopies($flag) {
    $base = <error descr="Malformed DateInterval specification.">'2 days'</error>;
    $mid = $flag ? $base : ($base);
    $s = $mid ?: $base;
    $s = $s ?? $mid;
    return new DateInterval($s);
}
function chained() {
    $a = $b = <error descr="Malformed DateInterval specification.">'P1'</error>;
    return new DateInterval($a);
}
