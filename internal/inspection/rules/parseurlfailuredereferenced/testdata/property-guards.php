<?php
// A property tested before it is read holds when the read runs no code and nothing between may change it.
final class F { public $p; }
class T { public $p; }
class Own { private $p; function own($n) { $this->p = parse_url($n); if ($this->p) { echo $this->p['host']; } } function changed($n) { $this->p = parse_url($n); if ($this->p) { $this->reset(); echo <warning descr="Reject URL parsing failure before indexing.">$this->p['host']</warning>; } } function reset() { $this->p = false; } }
function finalTruth(F $o, $n) { $o->p = parse_url($n); if ($o->p) { echo $o->p['host']; } }
function finalWrite(F $o, $n) { $o->p = parse_url($n); if ($o->p) { $o->p = parse_url($n . 'x'); echo $o->p['host']; } }
function finalOther(F $o, F $q, $n) { $o->p = parse_url($n); if ($o->p) { $q->p = false; echo <warning descr="Reject URL parsing failure before indexing.">$o->p['host']</warning>; } }
function openTruth(T $o, $n) { $o->p = parse_url($n); if ($o->p) { echo $o->p['host']; } }
function finalCall(F $o, $n) { $o->p = parse_url($n); if ($o->p) { work($o); echo <warning descr="Reject URL parsing failure before indexing.">$o->p['host']</warning>; } }
