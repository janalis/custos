<?php
function upgradeInCondition($path) {
    $h = fopen($path, 'c+');
    flock($h, LOCK_SH);
    if (flock($h, LOCK_EX)) {
        fwrite($h, 'x');
    }
}
function upgradeWithGuard($path) {
    $h = fopen($path, 'c+');
    flock($h, LOCK_SH);
    if (!flock($h, LOCK_EX)) { throw new RuntimeException('busy'); }
    ftruncate($h, 0);
}
function conditionalUpgrade($path, $w) {
    $h = fopen($path, 'c+');
    flock($h, LOCK_SH);
    if ($w) { flock($h, LOCK_EX); }
    if ($w) { fwrite($h, 'x'); }
}
function shortCircuitUpgrade($path) {
    $h = fopen($path, 'c+');
    flock($h, LOCK_SH);
    $ok = flock($h, LOCK_EX | LOCK_NB) || flock($h, LOCK_EX);
    fputs($h, 'x');
}
function aliasUpgrade($path) {
    $h = fopen($path, 'c+');
    flock($h, LOCK_SH);
    $x = $h;
    if (flock($x, LOCK_EX)) {
        fwrite($h, 'x');
    }
}
function maybeAliasUpgrade($path, $other, $c) {
    $h = fopen($path, 'c+');
    flock($h, LOCK_SH);
    $x = $c ? $h : $other;
    if (flock($x, LOCK_EX)) {
        fwrite($h, 'x');
    }
}
function stillShared($path) {
    $h = fopen($path, 'c+');
    flock($h, LOCK_SH);
    if (flock($h, LOCK_SH | LOCK_NB)) {
        <warning descr="Acquire an exclusive lock before writing.">fwrite($h, 'x')</warning>;
    }
}
function otherStreamUpgraded($path) {
    $h = fopen($path, 'c+');
    $o = fopen($path . '.lock', 'c+');
    flock($h, LOCK_SH);
    if (flock($o, LOCK_EX)) {
        <warning descr="Acquire an exclusive lock before writing.">fwrite($h, 'x')</warning>;
    }
}
function lockAfterWrite($path) {
    $h = fopen($path, 'c+');
    flock($h, LOCK_SH);
    strlen($path);
    <warning descr="Acquire an exclusive lock before writing.">fwrite($h, 'x')</warning>;
    flock($h, LOCK_EX);
}
