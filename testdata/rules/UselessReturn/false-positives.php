<?php
function counter() {
    static $n = 0;
    return $n = $n + 1;
}

function guarded() {
    try {
        return $res = compute();
    } finally {
        release($res);
    }
}

function wrapped() {
    return ($tmp = compute());
}

function compound($v) {
    return $v .= 'x';
}

function glob() {
    global $cfg;
    return $cfg = load();
}

function shared() {
    $later = function () use (&$val) { return $val; };
    return $val = 1;
}

function props($o) {
    if ($o) {
        return;
    }
    return $o->p = 1;
}

return $top = 1;
