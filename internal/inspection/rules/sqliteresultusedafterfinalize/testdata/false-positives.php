<?php
function reassigned(SQLite3Result $r, SQLite3Stmt $s) { $r->finalize(); $r = $s->execute(); $r->fetchArray(); }
function maybeReassigned(SQLite3Result $r, $c) { $r->finalize(); if ($c) { $r = other(); } $r->fetchArray(); }
function perIteration(SQLite3 $db, $qs) { foreach ($qs as $q) { $r = $db->query($q); $r->fetchArray(); $r->finalize(); } }
function conditional(SQLite3Result $r, $c) { if ($c) { $r->finalize(); } $r->fetchArray(); }
function shortCircuit(SQLite3Result $r, $c) { $c && $r->finalize(); $r->fetchArray(); }
function matchArm(SQLite3Result $r, $a) { match ($a) { 1 => $r->finalize(), default => null }; $r->fetchArray(); }
function nullSafeArgument(SQLite3Result $r, ?DateTime $d) { $d?->format($r->finalize()); $r->fetchArray(); }
function nullSafeProperty(SQLite3Result $r, $x) { $y = $x?->a->b[$r->finalize()]; $r->fetchArray(); }
function nullSafeCall(SQLite3Result $r, ?ArrayObject $x) { $y = $x?->getArrayCopy()[$r->finalize()]; $r->fetchArray(); }
function nullSafeName(SQLite3Result $r, $x) { $y = $x?->{$r->finalize()}; $r->fetchArray(); }
function inClosure(SQLite3Result $r) { $f = function () use ($r) { $r->finalize(); }; $r->fetchArray(); }
function inTry(SQLite3Result $r) { try { $r->finalize(); } catch (Exception $e) {} $r->fetchArray(); }
function otherResult(SQLite3Result $r, SQLite3Result $q) { $q->finalize(); $r->fetchArray(); }
function byReference(SQLite3Result &$r, SQLite3Result $q) { $r->finalize(); refresh(); $r->fetchArray(); $q->fetchArray(); }
function otherComparison(SQLite3Result $r, $c) { if ($c === true) { $r->reset(); } if ($r->finalize() === $c) { $r->reset(); } if ($r->finalize() !== true) { $r->reset(); } }
function elseBranch(SQLite3Result $r) { if ($r->finalize()) { } else { $r->fetchArray(); } }
function twice(SQLite3Result $r) { $r->finalize(); $r->finalize(); }
class Holder { public SQLite3Result $r; function run() { $this->r->finalize(); $this->refresh(); $this->r->fetchArray(); } }
