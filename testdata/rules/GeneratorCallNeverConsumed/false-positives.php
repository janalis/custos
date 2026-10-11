<?php
function gen() { log_it('start'); yield 1; }
function log_it($m) { echo $m, "\n"; }
// First-class callable syntax builds a Closure and calls nothing.
gen(...);
// Writes to plain locals are not observable outside the generator.
function local() { $x = 1; yield $x; }
local();
function tally(array $rows) { $n = 0; $seen = []; foreach ($rows as $row) { $n++; $seen[$row] = true; yield $n; } }
tally([]);
// A by-value parameter is a local copy.
function scaled($factor) { $factor *= 2; yield $factor; }
scaled(2);
