<?php
function forwardComposedStream($record) { return $record; }
function wrapComposedStream($record) { return forwardComposedStream($record); }
function composedStreamCaller() {
    $stream = tmpfile();
    $record = ['stream' => $stream, 'alias' => $stream];
    $copy = wrapComposedStream($record);
    fclose($copy['stream']);
    <error descr="Use an open stream for this operation.">fread($record['alias'], 1)</error>;
}
function opaqueComposedStream($record) { opaque($record); return forwardComposedStream($record); }
function refusedComposedStreamCaller() {
    $record = ['stream' => tmpfile()];
    $copy = opaqueComposedStream($record);
    fclose($copy['stream']);
    fread($record['stream'], 1);
}
