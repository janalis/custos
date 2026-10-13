<?php
// A stepped position is no longer the literal it was initialised with.
function post_increment(SQLite3Stmt $s) { $i = 0; $i++; $s->bindValue($i, 1); }
function pre_increment(SQLite3Stmt $s) { $i = 0; ++$i; $s->bindValue($i, 1); }
function decrement(SQLite3Stmt $s) { $i = 0; $i--; $s->bindParam(($i), $v); }
function loop_counter(SQLite3Stmt $s, array $vals) { $i = 0; foreach ($vals as $v) { $i++; $s->bindValue($i, $v); } }
function loop_tail(SQLite3Stmt $s, array $vals) { $i = 0; foreach ($vals as $v) { $s->bindValue($i, $v); $i++; } }
function loop_condition(SQLite3Stmt $s) { $i = 0; do { $s->bindValue($i, 1); } while (++$i < 3); }
function for_counter(SQLite3Stmt $s) { for ($i = 0; $i < 3; $i++) { $s->bindValue($i, 1); } }
function while_counter(SQLite3Stmt $s) { $i = 0; while ($i < 3) { $s->bindValue($i, 1); ++$i; } }

// A closure holding the variable by reference may rewrite it.
function reference_capture(SQLite3Stmt $s) { $i = 0; $f = function () use (&$i) { $i = 3; }; $f(); $s->bindValue($i, 1); }

// Other writes already invalidate the literal.
function compound(SQLite3Stmt $s) { $i = 0; $i += 1; $s->bindValue($i, 1); }
function reassigned(SQLite3Stmt $s) { $i = 0; $i = $i + 1; $s->bindValue($i, 1); }
function inline_step(SQLite3Stmt $s) { $i = 0; $s->bindValue(++$i, 1); }
function named(SQLite3Stmt $s) { $s->bindValue(':name', 1); }
