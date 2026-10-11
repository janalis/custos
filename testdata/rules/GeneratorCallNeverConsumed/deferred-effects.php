<?php
function announce() { echo "started\n"; yield 1; }
<warning descr="Consume the generator to execute its body.">announce()</warning>;
function printed() { print "started\n"; yield 1; }
<warning descr="Consume the generator to execute its body.">printed()</warning>;
function loads() { require __DIR__ . '/boot.php'; yield 1; }
<warning descr="Consume the generator to execute its body.">loads()</warning>;
function evaluates($code) { eval($code); yield 1; }
<warning descr="Consume the generator to execute its body.">evaluates('')</warning>;
function builds() { yield new Connection(); }
<warning descr="Consume the generator to execute its body.">builds()</warning>;
function validates($v) { if (!$v) { throw new InvalidArgumentException('v'); } yield $v; }
<warning descr="Consume the generator to execute its body.">validates(0)</warning>;
function stops($v) { if ($v) { exit(1); } yield $v; }
<warning descr="Consume the generator to execute its body.">stops(0)</warning>;
function registry() { Registry::$count++; yield 1; }
<warning descr="Consume the generator to execute its body.">registry()</warning>;
function publishes() { $GLOBALS['ready'] = true; yield 1; }
<warning descr="Consume the generator to execute its body.">publishes()</warning>;
function fills(array &$out) { $out[] = 1; yield 1; }
<warning descr="Consume the generator to execute its body.">fills($buffer)</warning>;
function counts() { global $total; $total++; yield $total; }
<warning descr="Consume the generator to execute its body.">counts()</warning>;
function remembers() { static $calls = 0; $calls++; yield $calls; }
<warning descr="Consume the generator to execute its body.">remembers()</warning>;
function binds(Holder $h) { $slot = &$h->slot; $slot = 1; yield 1; }
<warning descr="Consume the generator to execute its body.">binds($holder)</warning>;
function rewrites(Holder $h) { foreach ($h->items as &$item) { $item = 0; } yield 1; }
<warning descr="Consume the generator to execute its body.">rewrites($holder)</warning>;
function template() { ?>started<?php yield 1; }
<warning descr="Consume the generator to execute its body.">template()</warning>;
