<?php
function addHeaderCarry($carry, $value) { return $carry + $value; }
$sum = array_reduce([1, -1], 'addHeaderCarry', 0);
$empty = array_filter([$_GET['header']], fn() => $sum);
foreach ($empty as $entry) { header($entry); }
function headerReducerFactory() { return fn($carry, $value) => $carry + $value; }
$returnedSum = array_reduce([1, -1], headerReducerFactory(), 0);
$empty = array_filter([$_GET['header']], fn() => $returnedSum);
foreach ($empty as $entry) { header($entry); }
function falsyHeaderPredicate($entry) { if ($entry) { return false; } return null; }
$empty = array_filter([$_GET['header']], 'falsyHeaderPredicate');
foreach ($empty as $entry) { header($entry); }
function defaultHeaderPredicate($unused, int $value = 0) { return $value; }
$empty = array_filter([$_GET['header']], 'defaultHeaderPredicate');
foreach ($empty as $entry) { header($entry); }
function headerPredicateFactory(int $value) { return fn() => $value; }
$empty = array_filter([$_GET['header']], headerPredicateFactory(0));
foreach ($empty as $entry) { header($entry); }
$kept = array_filter([$_GET['header']], headerPredicateFactory(1));
foreach ($kept as $entry) { <error descr="Reject line breaks and validate this header value.">header($entry)</error>; }
function coerceHeaderPredicate(int $value) { return $value; }
function wrappedHeaderPredicate($value) { return coerceHeaderPredicate($value); }
$empty = array_filter(['00'], 'wrappedHeaderPredicate');
foreach ($empty as $entry) { header($entry); }
function mixedHeaderPredicate($entry) { if ($entry) { return false; } return true; }
$possible = array_filter([$_GET['header']], 'mixedHeaderPredicate');
foreach ($possible as $entry) { <error descr="Reject line breaks and validate this header value.">header($entry)</error>; }
function emittingFalsyHeaderPredicate($entry) { header($entry); if ($entry) { return false; } return null; }
<error descr="Reject line breaks and validate this header value.">array_filter([$_GET['header']], 'emittingFalsyHeaderPredicate')</error>;
