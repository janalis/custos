<?php
function wholeArm($x) { return match($x) { 1, 2 => 'low', <warning descr="Remove the repeated match condition.">2</warning>, <warning descr="Remove the repeated match condition.">1</warning> => 'never', default => 'other' }; }
function sameArm($x) { return match($x) { 1 => 'a', 2, <warning descr="Remove the repeated match condition.">2</warning> => 'b', default => 'other' }; }
function partly($x) { return match($x) { 1 => 'a', <warning descr="Remove the repeated match condition.">1</warning>, 3 => 'b', default => 'other' }; }
function unknown($x) { return match($x) { 1 => 'a', <warning descr="Remove the repeated match condition.">1</warning>, f() => 'b', default => 'other' }; }
function local($x) { $a = 1; return match($x) { 1 => 'one', <warning descr="Remove the repeated match condition.">$a</warning> => 'same', default => 'other' }; }
function lastArm($x) { return match($x) { 1, 2 => 'low', <warning descr="Remove the repeated match condition.">2</warning>, <warning descr="Remove the repeated match condition.">1</warning> => 'never' }; }
