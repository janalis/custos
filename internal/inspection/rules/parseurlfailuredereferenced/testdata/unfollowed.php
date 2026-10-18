<?php
// Code the scope evaluates leaves only the guards written around the read.
function thenGuard($raw, $code) { $parts = parse_url($raw); echo is_array($parts) ? $parts['host'] : ''; eval($code); }
function elseGuard($raw, $code) { $parts = parse_url($raw); echo $parts === false ? '' : $parts['host']; eval($code); }
function andGuard($raw, $code) { $parts = parse_url($raw); if (is_array($parts) && $parts['host'] === 'x') { echo 1; } eval($code); }
function orGuard($raw, $code) { $parts = parse_url($raw); if (!$parts || $parts['host'] === 'x') { echo 1; } eval($code); }
