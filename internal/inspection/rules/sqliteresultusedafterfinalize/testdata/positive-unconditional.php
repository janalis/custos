<?php
function direct(SQLite3Result $r) { $r->finalize(); <error descr="Read the SQLite result before finalizing it.">$r->fetchArray()</error>; }
function queried(SQLite3 $db) { $r = $db->query('select 1'); $r->finalize(); return <error descr="Read the SQLite result before finalizing it.">$r->fetchArray()</error>; }
function several(SQLite3 $db) { $r = $db->query('select 1'); $r->finalize(); <error descr="Read the SQLite result before finalizing it.">$r->numColumns()</error>; <error descr="Read the SQLite result before finalizing it.">$r->columnName(0)</error>; <error descr="Read the SQLite result before finalizing it.">$r->columnType(0)</error>; <error descr="Read the SQLite result before finalizing it.">$r->reset()</error>; }
function stored(SQLite3Result $r) { $ok = $r->finalize(); if ($ok) { <error descr="Read the SQLite result before finalizing it.">$r->numColumns()</error>; } }
function nested(SQLite3Result $r, $c) { $r->finalize(); if ($c) { <error descr="Read the SQLite result before finalizing it.">$r->fetchArray()</error>; } }
function aliased(SQLite3Result $r) { $r->finalize(); $q = $r; <error descr="Read the SQLite result before finalizing it.">$q->fetchArray()</error>; }
function identical(SQLite3Result $r) { if ($r->finalize() === true) { <error descr="Read the SQLite result before finalizing it.">$r->numColumns()</error>; } }
function yoda(SQLite3Result $r) { if (true == $r->finalize()) { <error descr="Read the SQLite result before finalizing it.">$r->reset()</error>; } }
function sameArm(SQLite3Result $r, $a) { $x = match ($a) { default => [$r->finalize(), <error descr="Read the SQLite result before finalizing it.">$r->fetchArray()</error>] }; return $x; }
function knownCall(SQLite3Result $r, DateTime $d) { $y = $d->format($r->finalize()); <error descr="Read the SQLite result before finalizing it.">$r->fetchArray()</error>; }
function arrayKey(SQLite3Result $r, $x) { $y = $x[0][$r->finalize()]; <error descr="Read the SQLite result before finalizing it.">$r->fetchArray()</error>; }
function chained(SQLite3Result $r) { $y = $r->finalize()->a; <error descr="Read the SQLite result before finalizing it.">$r->fetchArray()</error>; }
function sameChain(SQLite3Result $r, ?DateTime $d) { $y = $d?->format([$r->finalize(), <error descr="Read the SQLite result before finalizing it.">$r->numColumns()</error>]); }
