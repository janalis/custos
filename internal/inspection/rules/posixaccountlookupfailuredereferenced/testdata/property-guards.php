<?php
// A property tested before it is read holds when the read runs no code and nothing between may change it.
final class F { public $p; }
class T { public $p; }
class Own { private $p; function own($n) { $this->p = posix_getpwnam($n); if ($this->p) { echo $this->p['dir']; } } function changed($n) { $this->p = posix_getpwnam($n); if ($this->p) { $this->reset(); echo <warning descr="Reject account lookup failure before indexing.">$this->p['dir']</warning>; } } function reset() { $this->p = false; } }
function finalTruth(F $o, $n) { $o->p = posix_getpwnam($n); if ($o->p) { echo $o->p['dir']; } }
function finalWrite(F $o, $n) { $o->p = posix_getpwnam($n); if ($o->p) { $o->p = posix_getpwnam($n . 'x'); echo $o->p['dir']; } }
function finalOther(F $o, F $q, $n) { $o->p = posix_getpwnam($n); if ($o->p) { $q->p = false; echo <warning descr="Reject account lookup failure before indexing.">$o->p['dir']</warning>; } }
function openTruth(T $o, $n) { $o->p = posix_getpwnam($n); if ($o->p) { echo $o->p['dir']; } }
function finalCall(F $o, $n) { $o->p = posix_getpwnam($n); if ($o->p) { work($o); echo <warning descr="Reject account lookup failure before indexing.">$o->p['dir']</warning>; } }
