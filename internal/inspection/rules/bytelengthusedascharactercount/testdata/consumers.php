<?php
namespace App;

function substr($s, $n) { return $s; }

// A same-named namespaced function is not the byte-oriented builtin.
substr('x', <warning descr="Count Unicode characters with the intended counting model.">strlen('é')</warning>);
// Not a byte operation.
str_repeat('x', <warning descr="Count Unicode characters with the intended counting model.">strlen('é')</warning>);
// Comparison, concatenation and logical negation do not carry a byte quantity.
\substr('x', <warning descr="Count Unicode characters with the intended counting model.">strlen('é')</warning> > 1);
\fwrite(STDOUT, 'n=' . <warning descr="Count Unicode characters with the intended counting model.">strlen('é')</warning>);
\str_pad('x', !<warning descr="Count Unicode characters with the intended counting model.">strlen('é')</warning>);
$obj->substr('x', <warning descr="Count Unicode characters with the intended counting model.">strlen('é')</warning>);
$n = -<warning descr="Count Unicode characters with the intended counting model.">strlen('é')</warning> + 1;
