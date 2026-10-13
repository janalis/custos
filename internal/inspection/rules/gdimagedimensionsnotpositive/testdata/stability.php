<?php
function otherCounter() { $n = 0; for ($i = 0; $i < 3; $i++) { } return <error descr="Supply positive image dimensions.">imagecreate($n, 5)</error>; }
function capturedByValue() { $n = 0; $g = function () use ($n) { $n++; }; return <error descr="Supply positive image dimensions.">imagecreate($n, 5)</error>; }
function arrowLocalWrite() { $n = 0; $f = fn() => $n++; return <error descr="Supply positive image dimensions.">imagecreate($n, 5)</error>; }
function copied() { $n = 0; $m = $n; $n = $m; load(); return <error descr="Supply positive image dimensions.">imagecreatetruecolor(5, $m)</error>; }
function sharedWithoutCalls(&$n) { $n = 0; return <error descr="Supply positive image dimensions.">imagecreate($n, 5)</error>; }
function parameterAlias(&$other, $p = null) { $n = 0; load(); return <error descr="Supply positive image dimensions.">imagecreate($n, 5)</error>; }
$unrelated = 1;
$keep = function () use (&$unrelated) { $n = -1; load(); return <error descr="Supply positive image dimensions.">imagecreate($n, 5)</error>; };
