<?php

class Invoice
{
    var <weak_warning descr="Declare the visibility of 'total' explicitly.">$total</weak_warning>;
    static <weak_warning descr="Declare the visibility of 'registry' explicitly.">$registry</weak_warning> = [];
    readonly int <weak_warning descr="Declare the visibility of 'number' explicitly.">$number</weak_warning>;
    var <weak_warning descr="Declare the visibility of 'a' explicitly.">$a</weak_warning>, <weak_warning descr="Declare the visibility of 'b' explicitly.">$b</weak_warning>;
    protected $lines;
    private(set) string $currency;

    function <weak_warning descr="Declare the visibility of 'render' explicitly.">render</weak_warning>() {}
    final static function <weak_warning descr="Declare the visibility of 'make' explicitly.">make</weak_warning>() {}
    #[Pure]
    function <weak_warning descr="Declare the visibility of 'pure' explicitly.">pure</weak_warning>() {}
    public function save() {}

    const <weak_warning descr="Declare the visibility of 'MAX_ROWS' explicitly.">MAX_ROWS</weak_warning> = 500,
          MIN_ROWS = 1;
    protected const PAGE = 20;
}

abstract class Shape
{
    abstract function <weak_warning descr="Declare the visibility of 'area' explicitly.">area</weak_warning>();
}

interface Printable
{
    function <weak_warning descr="Declare the visibility of 'printOut' explicitly.">printOut</weak_warning>();
}

trait Greets
{
    static function <weak_warning descr="Declare the visibility of 'hello' explicitly.">hello</weak_warning>() {}
}

$o = new class {
    function <weak_warning descr="Declare the visibility of 'run' explicitly.">run</weak_warning>() {}
};
