<?php
function streamRecordFactory() {
    $stream = tmpfile();
    return ['stream' => $stream, 'alias' => $stream];
}
function streamRecordCaller() {
    $record = streamRecordFactory();
    $other = streamRecordFactory();
    fclose($record['stream']);
    <error descr="Use an open stream for this operation.">fread($record['alias'], 1)</error>;
    fread($other['stream'], 1);
}
function borrowedRecord($stream) { return ['stream' => $stream]; }
function borrowedRecordCaller() {
    $stream = tmpfile();
    fclose($stream);
    $record = borrowedRecord($stream);
    <error descr="Use an open stream for this operation.">fread($record['stream'], 1)</error>;
}
function escapedRecordCaller() {
    $record = streamRecordFactory();
    fclose($record['stream']);
    opaqueRecordConsumer($record);
    fread($record['alias'], 1);
}
