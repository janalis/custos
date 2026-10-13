<?php
function pushed() { $ids = [1, 2]; array_push($ids, '7', '2'); return array_keys(array_count_values($ids)); }
function mapped() { $ids = [1, 2]; foreach ($ids as &$id) { $id = "$id"; } return array_keys(array_count_values($ids)); }
function aliased() { $n = [1, 2]; $r = &$n; $r[] = 1.5; return count(array_count_values($n)); }
function captured() { $n = [1, 2]; $f = function () use (&$n) { $n = [1.5, 1.7]; }; $f(); return count(array_count_values($n)); }
function walked() { $n = [1, 2]; array_walk($n, function (&$v) { $v /= 4; }); return count(array_count_values($n)); }
function handed(Sink $sink) { $n = [1, 2]; $sink->take($n); Sink::keep($n); new Sink($n); return count(array_count_values($n)); }
function variadic(int ...$n) { array_push($n, 1.5); return count(array_count_values($n)); }
function shared(&$n, $code) { $n = [1, 2]; return count(array_count_values($n)); }
function merged(array $row) { $n = [1, 2]; extract($row); return count(array_count_values($n)); }
function loaded() { $n = [1, 2]; include 'more.php'; return count(array_count_values($n)); }
function evaluated($code) { $n = [1, 2]; eval($code); return count(array_count_values($n)); }
function target($rows) { $n = [1, 2]; foreach ($rows as $n[0]) {} return count(array_count_values($n)); }
function item() { $n = [1, 2]; $box = [&$n]; $box[0][] = 1.5; return count(array_count_values($n)); }
