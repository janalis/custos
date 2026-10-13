<?php
// Restored at the end of the body, before the next selection.
function tail($a, $b, $running) {
    $read = [$a, $b]; $write = null; $except = null;
    while ($running) {
        stream_select($read, $write, $except, 1);
        foreach ($read as $s) { fread($s, 10); }
        $read = [$a, $b];
    }
}
// Restored by the loop condition.
function cond($a, $b, $all) {
    $read = [$a, $b]; $write = null; $except = null;
    while (($read = $all) && stream_select($read, $write, $except, 1)) {
        echo "x";
    }
}
// Restored through destructuring.
function shortList($a, $b, $running, $o) {
    $read = [$a, $b]; $write = null; $except = null;
    while ($running) {
        [$read, $write] = $o->sets();
        stream_select($read, $write, $except, 1);
    }
}
function longList($a, $b, $running, $o) {
    $read = [$a, $b]; $write = null; $except = null;
    while ($running) {
        list(, list($read)) = $o->sets();
        stream_select($read, $write, $except, 1);
    }
}
// Rebound by a foreach target.
function eachValue($a, $b, $sets) {
    $read = [$a, $b]; $write = null; $except = null;
    while (true) {
        foreach ($sets as $read) {
            stream_select($read, $write, $except, 1);
        }
    }
}
function eachKey($a, $b, $sets) {
    $write = [$a, $b]; $read = null; $except = null;
    while (true) {
        foreach ($sets as $write => $ignored) {
            stream_select($read, $write, $except, 1);
        }
    }
}
// Handed to code that may refill it by reference.
function unknownFunction($a, $b, $running) {
    $read = [$a, $b]; $write = null; $except = null;
    while ($running) {
        rebuild_watch($read);
        stream_select($read, $write, $except, 1);
    }
}
function method($a, $b, $running, $o) {
    $except = [$a, $b]; $read = null; $write = null;
    while ($running) {
        stream_select($read, $write, $except, 1);
        $o->refill($except);
    }
}
