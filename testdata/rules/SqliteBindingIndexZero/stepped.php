<?php
// The position still holds its literal zero when the call runs.
function later_step(SQLite3Stmt $s) { $i = 0; <error descr="Start SQLite binding positions at one.">$s->bindValue($i, 1)</error>; $i++; }
function later_capture(SQLite3Stmt $s) { $i = 0; <error descr="Start SQLite binding positions at one.">$s->bindValue($i, 1)</error>; $f = function () use (&$i) { $i = 3; }; }
function value_capture(SQLite3Stmt $s) { $i = 0; $f = function () use ($i) { return $i; }; <error descr="Start SQLite binding positions at one.">$s->bindParam($i, $f)</error>; }
function step_after_loop(SQLite3Stmt $s, array $vals) { $i = 0; foreach ($vals as $v) { <error descr="Start SQLite binding positions at one.">$s->bindValue($i, $v)</error>; } $i++; }
