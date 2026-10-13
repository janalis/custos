<?php

class Invoice
{
    public $total;
    public static $registry = [];
    public readonly int $number;
    public $a, $b;
    protected $lines;
    private(set) string $currency;

    public function render() {}
    final public static function make() {}
    #[Pure]
    public function pure() {}
    public function save() {}

    public const MAX_ROWS = 500,
          MIN_ROWS = 1;
    protected const PAGE = 20;
}

abstract class Shape
{
    abstract public function area();
}

interface Printable
{
    public function printOut();
}

trait Greets
{
    public static function hello() {}
}

$o = new class {
    public function run() {}
};
