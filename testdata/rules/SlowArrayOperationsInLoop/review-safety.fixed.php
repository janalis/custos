<?php
function halve(string $rest): array { return [substr($rest, 0, 1), substr($rest, 1)]; }
function inspect($value): void {}

class Registry
{
    private static array $pending = [];
    private static int $at = 0;
    private array $items = [];
    private int $k = 0;

    public static function take(): string { return array_pop(self::$pending); }

    public function statics(Registry $other): void
    {
        for ($i = 0; $i < count(self::$pending); $i++) { self::take(); }
        for ($i = 0; $i < count(self::$pending); $i++) { $other->shrink(); }
        for ($i = 0; $i < count(self::$pending); $i++) { new Registry(); }
        for ($i = 0; $i < count(self::$pending); $i++) { new Registry; }
        for ($i = 0; $i < count(self::$pending); $i++) { inspect($i); }
        for ($i = 0; $i < count(self::$pending); $i++) { mystery(); }
        for ($i = 0; $i < count(self::$pending); $i++) { $f = fn () => self::take(); }
        for ($i = 0, $iMax = count(self::$pending); $i < $iMax; $i++) { echo strlen('x'); }
    }

    public function holders(SplQueue $queue, ArrayObject $todo, $node, array $nodes, int $k, array $a, string $s): void
    {
        for ($i = 0; $i < count($queue); $i++) { $queue->dequeue(); }
        for ($i = 0; $i < count($todo); $i++) { $todo?->append(1); }
        for ($i = 0; $i < count($queue); $i++) { inspect($queue); }
        for ($i = 0, $iMax = count($a); $i < $iMax; $i++) { inspect($a); }
        for ($i = 0, $iMax = strlen($s); $i < $iMax; $i++) { inspect($s); }
        for ($i = 0, $iMax = count($a[$k]); $i < $iMax; $i++) { inspect($k); }
        for ($i = 0; $i < count($node->children); $i++) { $node = $node->next; }
        for ($i = 0; $i < count($node->children); $i++) { inspect($node); }
        for ($i = 0; $i < count($node->children); $i++) { self::take(); }
        for ($i = 0; $i < count($nodes[$k]->children); $i++) { $nodes[$k]->grow(); }
        for ($i = 0; $i < count($node->list->children); $i++) { $node->refresh(); }
        for ($i = 0; $i < count($this->items); $i++) { self::shrink(); }
        for ($i = 0; $i < count($this->items); $i++) { Registry::take(); }
        for ($i = 0; $i < count($this->items); $i++) { $f = function () { $this->items[] = 1; }; }
        for ($i = 0; $i < count($a[$this->k]); $i++) { $this->k = 2; }
        for ($i = 0; $i < count($a[$k + 1]); $i++) { $k--; }
        for ($i = 0; $i < count($a[self::$at]); $i++) { self::$at++; }
    }

    public function writes(string $rest, array $a, array $q): void
    {
        for ($i = 0; $i < strlen($rest); $i++) { [$head, $rest] = halve($rest); }
        for ($i = 0; $i < strlen($rest); $i++) { list(, $rest) = halve($rest); }
        for ($i = 0, $iMax = strlen($rest); $i < $iMax; $i++) { [$head, [, $tail]] = halve($rest); }
        for ($i = 0; $i < count($a); $i++) { foreach ($q as $a) {} }
        for ($i = 0; $i < count($a); $i++) { foreach ($q as $a => $v) {} }
        for ($i = 0, $iMax = count($a); $i < $iMax; $i++) { foreach ($q as $key => $v) {} }
        for ($i = 0; array_pop($q), $i < count($q); $i++) {}
        for ($i = 0; $i < count($a); $i++) { sscanf('1', '%d', vars: $a); }
        for ($i = 0, $iMax = count($a); $i < $iMax; $i++) { echo in_array(needle: $i, haystack: $a); }
    }

    public function referenced(array $g, array $h): void
    {
        $ref = &$g;
        for ($i = 0; $i < count($g); $i++) { array_pop($ref); }
        $h = &$this->items;
        for ($i = 0; $i < count($h); $i++) { $this->shrink(); }
    }

    public function captured(array $g): void
    {
        $f = function () use (&$g) { $g[] = 1; };
        for ($i = 0; $i < count($g); $i++) { $f(); }
    }

    public function rows(array $grid, int $r): void
    {
        foreach ($grid as &$row) {}
        for ($i = 0; $i < count($grid[$r]); $i++) { $row[] = 1; }
    }

    public function harmless(array $g, array $list): void
    {
        $first = &$g[0];
        foreach ($list as &$item) {}
        $h = fn () => 1;
        $o = new class { public function m(): void {} };
        $c = function (array $inner) use ($g) {
            for ($i = 0, $iMax = count($inner); $i < $iMax; $i++) {}
        };
        for ($i = 0, $iMax = count($g); $i < $iMax; $i++) { $first = $i; }
        for ($i = 0, $iMax = count($list); $i < $iMax; $i++) { $item = $i; }
        for ($i = 0, $iMax = count($g); $i < $iMax; $i++) { $o = new class { public function m(array $g): void { $g[] = 1; } }; }
    }

    private function shrink(): void { array_pop($this->items); }
}

for ($i = 0, $iMax = count($argv); $i < $iMax; $i++) {}
