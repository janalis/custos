<?php
// The destructured value has no other use: the header form is equivalent.
function plain(array $rows) {
    $row = null;
    foreach ($rows as $row) { <weak_warning descr="Destructure directly in the foreach header.">[$a, $b] = $row</weak_warning>; echo $a, $b; }
}
function altSyntax(array $rows) {
    foreach ($rows as $key => $row):
        <weak_warning descr="Destructure directly in the foreach header.">list($a, , $b) = $row</weak_warning>;
        echo $key, $a, $b;
    endforeach;
}
function reusedName(array $rows, array $more) {
    foreach ($rows as $row) { <weak_warning descr="Destructure directly in the foreach header.">[$a, $b] = $row</weak_warning>; }
    foreach ($more as $row) { echo $row; }          // assigns $row again before reading it
    foreach ($more as $row => $unused) { echo $row; }
}
function nestedLoops(array $groups) {
    foreach ($groups as $rows) { foreach ($rows as $row) { <weak_warning descr="Destructure directly in the foreach header.">[$a, $b] = $row</weak_warning>; } }
}
function namesBefore(array $rows) {
    $all = compact('rows');
    foreach ($rows as $row) { <weak_warning descr="Destructure directly in the foreach header.">list('a' => $a) = $row</weak_warning>; }
    return $all;
}
function innerReuse(array $rows) {
    foreach ($rows as $row) { <weak_warning descr="Destructure directly in the foreach header.">[$a, $b] = $row</weak_warning>; foreach ($a as $row) { echo $row; } }
}
function inClosure(array $rows) {
    return function () use ($rows) { foreach ($rows as $row) { <weak_warning descr="Destructure directly in the foreach header.">[$a, $b] = $row</weak_warning>; } };
}
function propertyTarget(array $rows, $o) {
    foreach ($rows as [$o->first, $rest]) { <weak_warning descr="Destructure directly in the foreach header.">[$a, $b] = $rest</weak_warning>; }
}

// D4c: the value is still needed, so it must stay a variable.
function readAfterStatement(array $rows, array $seen) { foreach ($rows as $row) { [$a, $b] = $row; $seen[$a] = $row; } return $seen; }
function readAfterLoop(array $rows) { foreach ($rows as $row) { [$a, $b] = $row; } return $row; }
function interpolated(array $rows) { foreach ($rows as $row) { [$a, $b] = $row; echo "v $row[0]"; } }
function arrowCapture(array $rows) { foreach ($rows as $row) { [$a, $b] = $row; $f = fn() => $row; } }
function closureCapture(array $rows) { foreach ($rows as $row) { [$a, $b] = $row; $f = function () use ($row) {}; } }
function namesAfter(array $rows) { foreach ($rows as $row) { [$a, $b] = $row; } return compact('a'); }
function allVars(array $rows) { foreach ($rows as $row) { [$a, $b] = $row; } return get_defined_vars(); }
function extracted(array $rows) { foreach ($rows as $row) { [$a, $b] = $row; extract($a); } }
function variableVariable(array $rows) { foreach ($rows as $row) { [$a, $b] = $row; echo $$a; } }
function included(array $rows) { foreach ($rows as $row) { [$a, $b] = $row; } include 'tail.php'; }
function evaluated(array $rows) { foreach ($rows as $row) { [$a, $b] = $row; } eval('echo $row[0];'); }
function isGlobal(array $rows) { global $row; foreach ($rows as $row) { [$a, $b] = $row; } }
function isStatic(array $rows) { static $row; foreach ($rows as $row) { [$a, $b] = $row; } }
function refParam(array $rows, &$row) { foreach ($rows as $row) { [$a, $b] = $row; } }
function refUse(array $rows) { return function () use (&$row, $rows) { foreach ($rows as $row) { [$a, $b] = $row; } }; }
function outerLoop(array $rows) { while (more()) { echo $row; foreach ($rows as $row) { [$a, $b] = $row; } } }
function laterSubject(array $rows) { foreach ($rows as $row) { [$a, $b] = $row; } foreach ($row as $row) {} }
function laterBody(array $rows) { foreach ($rows as $row) { [$a, $b] = $row; } foreach ($rows as $k => $z) { echo $row; } }

// D4: only a by-value foreach value can become a list.
function keyVariable(iterable $map) { foreach ($map as $key => $value) { [$type, $id] = $key; } }
function byReference(array $rows) { foreach ($rows as &$row) { [$a, $b] = $row; $row[] = $a . $b; } }

// D4b: targets that would be evaluated differently in the header.
function keyAsTarget(array $rows) { foreach ($rows as $k => $row) { [$a, $k] = $row; } }
function keyInTarget(array $rows, array $out) { foreach ($rows as $k => $row) { [$out[$k], $b] = $row; } return $out; }
function refTarget(array $rows) { foreach ($rows as $row) { [&$a, $b] = $row; } }
function dynamicTarget(array $rows) { foreach ($rows as $row) { [$a, $$b] = $row; } }
