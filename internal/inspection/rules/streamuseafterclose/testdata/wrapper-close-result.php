<?php
function closeWithResult($stream): bool { return fclose($stream); }
function closeWithIntegerResult($stream): int { return fclose($stream); }
function closeResultCaller() {
    $stream = tmpfile();
    if ($stream === false) { return; }
    $closed = closeWithResult($stream);
    if ($closed === true) {
        <error descr="Use an open stream for this operation.">fread($stream, 1)</error>;
    }
}
function failedCloseResultCaller() {
    $stream = tmpfile();
    if ($stream === false) { return; }
    if (!closeWithResult($stream)) { fread($stream, 1); }
}
function coercedCloseResultCaller() {
    $stream = tmpfile();
    if ($stream === false) { return; }
    if (closeWithIntegerResult($stream) !== false) { fread($stream, 1); }
}
