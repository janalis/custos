<?php
// A fresh array value has no other observer: writing through the leftover
// reference cannot reach any data the program still holds.
function freshRows() { return [['n' => 1], ['n' => 2]]; }
function typedRows(): array { return [1, 2]; }
class FreshSource
{
    public $items = [1, 2];
    public function rows(): array { return $this->items; }
    public static function all(): array { return [1]; }
}
function fromTemporaries(FreshSource $source, array $input)
{
    foreach (freshRows() as &$row) { $row['n'] *= 2; }
    $row = null;
    foreach ((typedRows()) as &$typed) {}
    $typed = 0;
    foreach ([1, 2, 3] as &$literal) { $literal++; }
    $literal = 0;
    foreach (array(1, 2) as &$long) {}
    $long = 0;
    foreach ($source->rows() as &$fromMethod) {}
    $fromMethod = 0;
    foreach (FreshSource::all() as &$fromStatic) {}
    $fromStatic = 0;
    foreach (array_values($input) as &$fromBuiltin) {}
    $fromBuiltin = 0;
}
