<?php
function wholeArm($x) { return match($x) { 1, 2 => 'low',  default => 'other' }; }
function sameArm($x) { return match($x) { 1 => 'a', 2, 2 => 'b', default => 'other' }; }
function partly($x) { return match($x) { 1 => 'a', 1, 3 => 'b', default => 'other' }; }
function unknown($x) { return match($x) { 1 => 'a', 1, f() => 'b', default => 'other' }; }
function local($x) { $a = 1; return match($x) { 1 => 'one',  default => 'other' }; }
function lastArm($x) { return match($x) { 1, 2 => 'low' }; }
