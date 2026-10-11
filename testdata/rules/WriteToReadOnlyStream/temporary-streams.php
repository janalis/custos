<?php
function memoryReadOnly() { $buffer = fopen('php://memory', 'r'); if ($buffer === false) { return; } <error descr="Open the stream with write access.">fwrite($buffer, "draft")</error>; }
function memoryBinaryRead() { $buffer = fopen('php://memory', 'rb'); if ($buffer === false) { return; } <error descr="Open the stream with write access.">fputs($buffer, "draft")</error>; }
function memoryExclusive() { $buffer = fopen('php://memory', 'x'); if ($buffer === false) { return; } <error descr="Open the stream with write access.">fwrite($buffer, "draft")</error>; }
function tempCreate() { $spool = fopen('php://temp', 'c'); if ($spool === false) { return; } <error descr="Open the stream with write access.">ftruncate($spool, 0)</error>; }
function plainFileReadWrite($path) { $file = fopen($path, 'rw'); if ($file === false) { return; } <error descr="Open the stream with write access.">fwrite($file, "draft")</error>; }
