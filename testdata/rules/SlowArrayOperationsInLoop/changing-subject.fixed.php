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
        for ($i = 0; $i < count($this->items); $i++) { $this->remove($i); }
        for ($i = 0; $i < count($grid[$r]); $i++) { $r++; }
        for ($i = 0; $i < count($a); $i++) { unset($a[$i]); }
        for ($i = 0; $i < count($a); $i++) { mystery($a); }
        for ($i = 0; $i < count($this->load()); $i++) {}
        foreach ($a as &$v) {}
        for ($i = 0; $i < count($a); $i++) { $x = &$a; foreach ($a as &$w) {} }
        for ($i = 0, $iMax = count($a); $i < $iMax; $i++) { echo in_array($i, $a); }
        for ($i = 0; $i < count($a); $i++) { sprintf('%d', ...$a); }
        for ($i = 0; $i < count($a); $i++) { sscanf('1', '%d', $a); }
        for ($i = 0, $iMax = count($a); $i < $iMax; $i++) { echo strlen($a[$i]); $f = function () use ($a) { $a[] = 1; }; }
    }

    public function shapes(array $a, array $args, string $n): void
    {
        for ($i = 0; $i < count(...$args); $i++) {}
        for ($i = 0; $i < count($$n); $i++) {}
        for ($i = 0; $i < count($a); $i++) { foreach ($a as &$v) {} }
        for ($i = 0; $i < count($a); $i++) { $f = in_array(...); sscanf('1', '%d', $x, $a); }
        for ($i = 0; $i < count($a); $i++) { Box::take($a); }
        for ($i = 0; $i < count($a); $i++) { new Box($a); }
    }

    private function remove(int $i): void { unset($this->items[$i]); }
    private function load(): array { return $this->items; }
}
