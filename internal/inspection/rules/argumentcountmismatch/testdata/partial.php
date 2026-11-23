<?php
namespace Shop;

// A partial application (PHP 8.6) checks the count when the closure is
// built: each `?` stands for an argument, and a trailing `...` leaves the
// remaining parameters to the closure's call.

function total($items, $tax) {}
function everything() { return func_get_args(); }
function many(...$values) {}

$a = <error descr="Pass the required argument $tax.">total(?)</error>;
$b = <error descr="Pass the required argument $items.">total(tax: ?)</error>;
$c = <error descr="Pass the required argument $length.">str_pad(?)</error>;
$d = strlen(?, <error descr="Remove the extra argument: strlen() accepts at most 1.">1</error>);
$e = total(?, ?, <error descr="Remove the extra argument: Shop\total() accepts at most 2.">?</error>);
$f = everything(<error descr="Remove the extra argument: Shop\everything() accepts at most 0.">?</error>);
$g = total(1, 2, <error descr="Remove the extra argument: Shop\total() accepts at most 2.">3</error>, ...);

// Valid partial applications.
$h = total(?, 0.2);
$i = total(?, ?);
$j = total(1, ...);
$k = total(...);
$l = total(?, ...);
$m = total(tax: ?, items: ?);
$n = str_pad(?, 4);
$o = many(?, ?, ?);
$p = everything(...);
