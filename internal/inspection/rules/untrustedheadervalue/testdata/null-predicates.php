<?php
function discardHeaderEntry($entry) {}
$empty = array_filter([$_GET['header']], 'discardHeaderEntry');
foreach ($empty as $entry) { header($entry); }
function emitAndDiscardHeader($entry) { header($entry); }
<error descr="Reject line breaks and validate this header value.">array_filter([$_GET['header']], 'emitAndDiscardHeader')</error>;
$sum = array_reduce([1, -1], fn($carry, $entry) => $carry + $entry, 0);
$emptyAfterReduction = array_filter([$_GET['header']], fn() => $sum);
foreach ($emptyAfterReduction as $entry) { header($entry); }
