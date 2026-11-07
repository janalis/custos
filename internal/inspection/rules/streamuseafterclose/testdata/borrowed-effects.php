<?php
function checkedBorrowedClose($stream) {
    if (!fclose($stream)) { exit; }
}
function uncheckedBorrowedClose($stream) { fclose($stream); }
function checkedBorrowedCaller() {
    $stream = tmpfile();
    if ($stream === false) { return; }
    checkedBorrowedClose($stream);
    <error descr="Use an open stream for this operation.">fread($stream, 1)</error>;
}
function uncheckedBorrowedCaller() {
    $stream = tmpfile();
    if ($stream === false) { return; }
    uncheckedBorrowedClose($stream);
    fread($stream, 1);
}
