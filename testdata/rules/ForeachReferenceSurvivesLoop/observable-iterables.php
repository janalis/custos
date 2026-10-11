<?php
// The leftover reference still points into data that other code can read.
function &sharedRows() { static $rows = [1, 2, 3]; return $rows; }
function &sharedStream() { static $rows = [1, 2, 3]; foreach ($rows as &$item) { yield $item; } }
function plainObject() { return new stdClass(); }
class SharedSource
{
    public $items = [1, 2];
    public function &rows() { return $this->items; }
    public function lazy(): iterable { return $this->items; }
    public static function &all() { static $all = [1]; return $all; }
}
function fromObservables(SharedSource $source, $untyped, $a, $b, array $input)
{
    foreach (sharedRows() as &$v1) {}
    <warning descr="Unset the foreach reference before reusing the variable.">$v1 = 99</warning>;
    foreach (sharedStream() as &$v2) {}
    <warning descr="Unset the foreach reference before reusing the variable.">$v2 = 99</warning>;
    foreach (plainObject() as &$v3) {}
    <warning descr="Unset the foreach reference before reusing the variable.">$v3 = 99</warning>;
    foreach (not_declared_anywhere() as &$v4) {}
    <warning descr="Unset the foreach reference before reusing the variable.">$v4 = 99</warning>;
    foreach ($source->rows() as &$v5) {}
    <warning descr="Unset the foreach reference before reusing the variable.">$v5 = 99</warning>;
    foreach ($untyped->rows() as &$v6) {}
    <warning descr="Unset the foreach reference before reusing the variable.">$v6 = 99</warning>;
    foreach ($source->lazy() as &$v7) {}
    <warning descr="Unset the foreach reference before reusing the variable.">$v7 = 99</warning>;
    foreach (SharedSource::all() as &$v8) {}
    <warning descr="Unset the foreach reference before reusing the variable.">$v8 = 99</warning>;
    foreach ([&$a, &$b] as &$v9) {}
    <warning descr="Unset the foreach reference before reusing the variable.">$v9 = 99</warning>;
    foreach ([...$input] as &$v10) {}
    <warning descr="Unset the foreach reference before reusing the variable.">$v10 = 99</warning>;
    foreach ($source->items as &$v11) {}
    <warning descr="Unset the foreach reference before reusing the variable.">$v11 = 99</warning>;
    foreach ($input as &$v12) {}
    <warning descr="Unset the foreach reference before reusing the variable.">$v12 = 99</warning>;
    foreach ($input ?: [1] as &$v13) {}
    <warning descr="Unset the foreach reference before reusing the variable.">$v13 = 99</warning>;
}
