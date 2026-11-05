<?php
function reduceHeaderWrapper($value) {
    array_reduce([$value], fn($carry, $entry) => header($entry), '');
}
<error descr="Reject line breaks and validate this header value.">reduceHeaderWrapper($_GET['header'])</error>;
$value = array_reduce(['safe', $_GET['header']], fn($carry, $entry) => $entry, '');
<error descr="Reject line breaks and validate this header value.">header($value)</error>;
$last = array_reduce([$_GET['header'], 'X-Safe: yes'], fn($carry, $entry) => $entry, '');
header($last);
function rejectAllHeaders($entry) { return false; }
$filtered = array_filter([$_GET['header']], 'rejectAllHeaders');
foreach ($filtered as $entry) { header($entry); }
