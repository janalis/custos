<?php
function guardedByExit($read, $write, $except) {
    $r = stream_select($read, $write, $except, 1);
    if ($r === false) { throw new RuntimeException('select failed'); }
    if ($r !== 0) { echo "ready"; }
}
function guardedByAnd($read, $write, $except) {
    $r = stream_select($read, $write, $except, 1);
    if ($r !== false && $r !== 0) { echo "ready"; }
}
function guardedByElseIf($read, $write, $except) {
    $r = stream_select($read, $write, $except, 1);
    if ($r === false) { return; } elseif ($r !== 0) { echo "ready"; }
}
function guardedByOr($read, $write, $except) {
    $r = stream_select($read, $write, $except, 1);
    if (false === $r || $r !== 0 && count($read)) { echo "handled"; }
}
function guardedByEnclosingTest($read, $write, $except) {
    $r = stream_select($read, $write, $except, 1);
    if ($r !== false) {
        if ($r !== 0) { echo "ready"; }
    }
}
function guardedByLooseExit($read, $write, $except) {
    $r = stream_select($read, $write, $except, 1);
    if ($r == false) { return; }
    if ($r !== 0) { echo "never reached with zero"; }
}
function guardedByTypeCheck($read, $write, $except) {
    $r = stream_select($read, $write, $except, 1);
    if (is_int($r) && $r !== 0) { echo "ready"; }
}
