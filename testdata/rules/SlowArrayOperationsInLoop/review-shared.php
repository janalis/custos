<?php
// Values that code called from the loop can reach: the length is reported,
// but hoisting it is only offered when the loop runs no such code.
function trimQueue(): void { global $queue; array_pop($queue); }

$queue = [1, 2, 3, 4];
for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($queue)</error>; $i++) { trimQueue(); }
for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($queue)</error>; $i++) { echo $queue[$i]; }

function viaGlobal(array $local, $cb): void
{
    global $queue;
    static $seen = [];
    for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($queue)</error>; $i++) { trimQueue(); }
    for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($seen)</error>; $i++) { viaGlobal($local, $cb); }
    for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($GLOBALS['queue'])</error>; $i++) { trimQueue(); }
    for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($_SESSION)</error>; $i++) { error_log((string) $i); }
    for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($_SESSION)</error>; $i++) { $cb(); }
    for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($_SESSION)</error>; $i++) { usort($local, $cb); }
    for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($_SESSION)</error>; $i++) { sort($local); }
    for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($local)</error>; $i++) { trimQueue(); }
    for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($local)</error>; $i++) { $f = function () { static $local; global $queue; }; }
    for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($local)</error>; $i++) { extract(['local' => []]); }
    for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($local)</error>; $i++) { $name = 'local'; $$name = []; }
    for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($local)</error>; $i++) { include 'more.php'; }
    for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($local)</error>; $i++) { eval('$local = [];'); }
}

function viaReference(array &$rows, array $copy): void
{
    for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($rows)</error>; $i++) { trimQueue(); }
    for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($copy)</error>; $i++) { trimQueue(); }
    $each = function () use (&$copy, $rows) {
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($copy)</error>; $i++) { trimQueue(); }
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($rows)</error>; $i++) { trimQueue(); }
    };
}

class Basket
{
    public array $items = [];
    public array $other = [];

    public function run(Basket $me, array $plain, $cb): \Generator
    {
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($this->items)</error>; $i++) { trimQueue(); }
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($this->items)</error>; $i++) { $cb(); }
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($this->items)</error>; $i++) { $plain[] = new Basket(); }
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($this->items)</error>; $i++) { Basket::reset(); }
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($this->items)</error>; $i++) { $plain[0]->touch(); }
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($this->items)</error>; $i++) { yield $i; }
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($this->items)</error>; $i++) { yield from $plain; }
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($this->items)</error>; $i++) { array_map($cb, $plain); }
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($this->items)</error>; $i++) { echo strlen($plain[$i]); }
        // The same property through another object expression.
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($this->items)</error>; $i++) { $me->items[] = $i; }
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($this->items)</error>; $i++) { unset($me->items[0], $plain[0]); }
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($this->items)</error>; $i++) { array_pop($me->items); }
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($this->items)</error>; $i++) { $me->other[] = $i; $plain[0] = $i; unset($me->other[0]); array_pop($me->other); echo count($me->items); }
    }
}
