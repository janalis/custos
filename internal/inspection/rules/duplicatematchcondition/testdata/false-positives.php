<?php
// A variable condition is trusted only when one plain assignment binds it.
function postIncrement($x) { $a = 1; $a++; return match($x) { 1 => 'one', $a => 'counter', default => 'other' }; }
function preDecrement($x) { $a = 1; --$a; return match($x) { 1 => 'one', $a => 'counter', default => 'other' }; }
function earlierArm($x) { $a = 1; return match($x) { 1 => 'one', ($a++) + 10 => 'eleven', $a => 'counter', default => 'other' }; }
function compound($x) { $a = 1; $a += 1; return match($x) { 1 => 'one', $a => 'two', default => 'other' }; }
function captured($x) { $a = 1; $bump = function () use (&$a) { $a = 2; }; $bump(); return match($x) { 1 => 'one', $a => 'bumped', default => 'other' }; }
function extracted($x) { $a = 1; extract(['a' => 2]); return match($x) { 1 => 'one', $a => 'extracted', default => 'other' }; }
function caught($x) { $a = 1; try { throw $x; } catch (Exception $a) {} return match($x) { 1 => 'one', $a => 'caught', default => 'other' }; }
function reassigned($x) { $a = 1; $a = 2; return match($x) { 2 => 'two', $a => 'again', default => 'other' }; }
function copied($x) { $b = 1; $b++; $a = $b; return match($x) { 1 => 'one', $a => 'copy', default => 'other' }; }
function destructured($x) { [$a] = [1]; return match($x) { 1 => 'one', $a => 'item', default => 'other' }; }
function aliased($x) { $b = 2; $a = &$b; return match($x) { 2 => 'two', $a => 'alias', default => 'other' }; }
function byName($x, $n) { $a = 1; $$n = 2; return match($x) { 1 => 'one', $a => 'named', default => 'other' }; }
function evaluated($x) { $a = 1; eval('$a = 2;'); return match($x) { 1 => 'one', $a => 'eval', default => 'other' }; }
function included($x) { $a = 1; include 'rebind.php'; return match($x) { 1 => 'one', $a => 'included', default => 'other' }; }
function persistent($x) { static $a = 1; return match($x) { 1 => 'one', $a => 'static', default => 'other' }; }
function parameter($x, $a) { return match($x) { 1 => 'one', $a => 'param', default => 'other' }; }
$arrow = fn($x) => match($x) { 1 => 'one', $x => 'self', default => 'other' };
$top = 1;
echo match($input) { 1 => 'one', $top => 'global', default => 'other' };
function dynamicName($x, $n) { return match($x) { 1 => 'one', $$n => 'named', default => 'other' }; }
