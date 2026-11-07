<?php
function indirectCloseGuard() {
    $stream = tmpfile();
    if ($stream === false) { return; }
    $close = fclose(...);
    if ($close($stream)) {
        <error descr="Use an open stream for this operation.">fread($stream, 1)</error>;
    }
}
function savedIndirectCloseGuard() {
    $stream = tmpfile();
    if ($stream === false) { return; }
    $ok = call_user_func('fclose', $stream);
    if ($ok) {
        <error descr="Use an open stream for this operation.">fread($stream, 1)</error>;
    }
}
function failedIndirectCloseGuard() {
    $stream = tmpfile();
    if ($stream === false) { return; }
    $close = fclose(...);
    if (!$close($stream)) {
        fread($stream, 1);
    }
}

function arrayIndirectCloseGuard() {
    $stream = tmpfile();
    if ($stream === false) { return; }
    if (call_user_func_array('fclose', [$stream])) {
        <error descr="Use an open stream for this operation.">fread($stream, 1)</error>;
    }
}
