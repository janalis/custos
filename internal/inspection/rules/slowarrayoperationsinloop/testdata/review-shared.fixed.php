<?php
// Values that code called from the loop can reach: the length is reported,
// but hoisting it is only offered when the loop runs no such code.
function trimQueue(): void { global $queue; array_pop($queue); }

$queue = [1, 2, 3, 4];
for ($i = 0; $i < count($queue); $i++) { trimQueue(); }
for ($i = 0, $iMax = count($queue); $i < $iMax; $i++) { echo $queue[$i]; }

function viaGlobal(array $local, $cb): void
{
    global $queue;
    static $seen = [];
    for ($i = 0; $i < count($queue); $i++) { trimQueue(); }
    for ($i = 0; $i < count($seen); $i++) { viaGlobal($local, $cb); }
    for ($i = 0; $i < count($GLOBALS['queue']); $i++) { trimQueue(); }
    for ($i = 0, $iMax = count($_SESSION); $i < $iMax; $i++) { error_log((string) $i); }
    for ($i = 0; $i < count($_SESSION); $i++) { $cb(); }
    for ($i = 0; $i < count($_SESSION); $i++) { usort($local, $cb); }
    for ($i = 0, $iMax = count($_SESSION); $i < $iMax; $i++) { sort($local); }
    for ($i = 0, $iMax = count($local); $i < $iMax; $i++) { trimQueue(); }
    for ($i = 0, $iMax = count($local); $i < $iMax; $i++) { $f = function () { static $local; global $queue; }; }
    for ($i = 0; $i < count($local); $i++) { extract(['local' => []]); }
    for ($i = 0; $i < count($local); $i++) { $name = 'local'; $$name = []; }
    for ($i = 0; $i < count($local); $i++) { include 'more.php'; }
    for ($i = 0; $i < count($local); $i++) { eval('$local = [];'); }
}

function viaReference(array &$rows, array $copy): void
{
    for ($i = 0; $i < count($rows); $i++) { trimQueue(); }
    for ($i = 0; $i < count($copy); $i++) { trimQueue(); }
    $each = function () use (&$copy, $rows) {
        for ($i = 0; $i < count($copy); $i++) { trimQueue(); }
        for ($i = 0, $iMax = count($rows); $i < $iMax; $i++) { trimQueue(); }
    };
}

class Basket
{
    public array $items = [];
    public array $other = [];

    public function run(Basket $me, array $plain, $cb): \Generator
    {
        for ($i = 0; $i < count($this->items); $i++) { trimQueue(); }
        for ($i = 0; $i < count($this->items); $i++) { $cb(); }
        for ($i = 0; $i < count($this->items); $i++) { $plain[] = new Basket(); }
        for ($i = 0; $i < count($this->items); $i++) { Basket::reset(); }
        for ($i = 0; $i < count($this->items); $i++) { $plain[0]->touch(); }
        for ($i = 0; $i < count($this->items); $i++) { yield $i; }
        for ($i = 0; $i < count($this->items); $i++) { yield from $plain; }
        for ($i = 0; $i < count($this->items); $i++) { array_map($cb, $plain); }
        for ($i = 0, $iMax = count($this->items); $i < $iMax; $i++) { echo strlen($plain[$i]); }
        // The same property through another object expression.
        for ($i = 0; $i < count($this->items); $i++) { $me->items[] = $i; }
        for ($i = 0; $i < count($this->items); $i++) { unset($me->items[0], $plain[0]); }
        for ($i = 0; $i < count($this->items); $i++) { array_pop($me->items); }
        for ($i = 0, $iMax = count($this->items); $i < $iMax; $i++) { $me->other[] = $i; $plain[0] = $i; unset($me->other[0]); array_pop($me->other); echo count($me->items); }
    }
}
