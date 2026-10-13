<?php
function memoryReadWrite() { $buffer = fopen('php://memory', 'rw'); if ($buffer === false) { return; } fwrite($buffer, "draft"); }
function tempReadWrite() { $spool = fopen('php://temp', 'rw'); if ($spool === false) { return; } fputs($spool, "draft"); }
function tempBinaryLimit() { $spool = fopen('php://temp/maxmemory:2048', 'rbw'); if ($spool === false) { return; } fputcsv($spool, array(1, 2)); }
function memoryReadAppend() { $buffer = fopen('php://memory', 'ra'); if ($buffer === false) { return; } ftruncate($buffer, 0); }
function memoryWriteRead() { $buffer = fopen('php://memory', 'wr'); if ($buffer === false) { return; } fwrite($buffer, "draft"); }
function unknownHandle($handle) { fwrite($handle, "draft"); }
