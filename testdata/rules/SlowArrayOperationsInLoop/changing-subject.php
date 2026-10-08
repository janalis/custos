<?php
class Box
{
    public function __construct(array $a) {}
    public static function take(array $a): void {}
}

class Queue
{
    private array $items = [];

    public function drain(array $a, string $s, array $grid, int $r): void
    {
        for ($i = 0; $i < count($a); $i++) { if ($a[$i] > 1) { array_pop($a); } }
        for ($i = 0; $i < count($a); $i++) { $a[] = 0; }
        for ($i = 0; $i < strlen($s); $i++) { $s = substr($s, 1); }
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($this->items)</error>; $i++) { $this->remove($i); }
        for ($i = 0; $i < count($grid[$r]); $i++) { $r++; }
        for ($i = 0; $i < count($a); $i++) { unset($a[$i]); }
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($a)</error>; $i++) { mystery($a); }
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($this->load())</error>; $i++) {}
        foreach ($a as &$v) {}
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($a)</error>; $i++) { $x = &$a; foreach ($a as &$w) {} }
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($a)</error>; $i++) { echo in_array($i, $a); }
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($a)</error>; $i++) { sprintf('%d', ...$a); }
        for ($i = 0; $i < count($a); $i++) { sscanf('1', '%d', $a); }
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($a)</error>; $i++) { echo strlen($a[$i]); $f = function () use ($a) { $a[] = 1; }; }
    }

    public function shapes(array $a, array $args, string $n): void
    {
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count(...$args)</error>; $i++) {}
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($$n)</error>; $i++) {}
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($a)</error>; $i++) { foreach ($a as &$v) {} }
        for ($i = 0; $i < count($a); $i++) { $f = in_array(...); sscanf('1', '%d', $x, $a); }
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($a)</error>; $i++) { Box::take($a); }
        for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($a)</error>; $i++) { new Box($a); }
    }

    public function grow(array $a): void
    {
        for ($i = 0; $i < count($this->items); ++$i) { $this->items['x' . $i] = 1; }
        for ($i = 0; $i < count($a); ++$i) { array_push($a, 1); }
        for ($i = 0; $i < count($a); ++$i) { if ($a[$i] === 0) { $a = []; } }
        for ($i = 0; $i < count(self::$all); ++$i) { self::$all[] = $i; }
    }

    private static array $all = [];

    private function remove(int $i): void { unset($this->items[$i]); }
    private function load(): array { return $this->items; }
}
