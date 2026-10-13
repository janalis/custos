<?php
// The method form of unset releases the slot.
$a = new SplFixedArray(4);
$a[3] = 'tail';
$a->offsetUnset(3);
$a->setSize(2);

// A closure holding the array may release slots before the shrink.
$b = new SplFixedArray(4);
$b[3] = 'tail';
$release = function () use ($b) { unset($b[3]); };
$release();
$b->setSize(2);

// So may an arrow function, an alias or an unknown callee.
$c = new SplFixedArray(4);
$c[3] = 'tail';
$drop = fn () => $c->offsetUnset(3);
$drop();
$c->setSize(2);

$d = new SplFixedArray(4);
$d[3] = 'tail';
$alias = $d;
unset($alias[3]);
$d->setSize(2);

$e = new SplFixedArray(4);
$e[3] = 'tail';
compact_entries($e);
$e->setSize(2);

// A release on only some paths leaves the occupancy unknown.
$f = new SplFixedArray(4);
$f[3] = 'tail';
if ($argc > 1) {
    unset($f[3]);
}
$f->setSize(2);
