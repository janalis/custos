<?php
function guardAfterTest($read, $write, $except) {
    $r = stream_select($read, $write, $except, 1);
    if (<warning descr="Handle selection failure before processing ready streams.">$r !== 0</warning> && $r !== false) { echo "ready"; }
}
function guardOnOtherResult($read, $write, $except, $other) {
    $r = stream_select($read, $write, $except, 1);
    if ($other !== false && <warning descr="Handle selection failure before processing ready streams.">$r !== 0</warning>) { echo "ready"; }
}
function guardBeforeNewSelect($read, $write, $except) {
    $r = stream_select($read, $write, $except, 1);
    if ($r === false) { return; }
    $r = stream_select($read, $write, $except, 1);
    if (<warning descr="Handle selection failure before processing ready streams.">$r !== 0</warning>) { echo "ready"; }
}
function directCalls($read, $write, $except) {
    if (stream_select($read, $write, $except, 1) !== false && <warning descr="Handle selection failure before processing ready streams.">stream_select($read, $write, $except, 1) !== 0</warning>) { echo "ready"; }
}
function laterArmTestsFailure($read, $write, $except, $flag) {
    $r = stream_select($read, $write, $except, 1);
    if ($flag) { return; } elseif (<warning descr="Handle selection failure before processing ready streams.">$r !== 0</warning>) { echo "ready"; } elseif ($r === false) { return; }
}
