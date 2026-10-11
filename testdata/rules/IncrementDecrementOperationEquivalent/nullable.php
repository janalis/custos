<?php
// '--' leaves null untouched while 'null - 1' is -1.
class Stock {
    private $left;
    private $start;
    private $none = null;
    private $zero = 0;
    private static $depth;
    private static $level = 0;
    private ?int $typed = null;
    public $a, $b = 1;

    public function __construct($start) { $this->start = $start; }

    public function take($o, $name, ?int $n, $u) {
        <weak_warning descr="Use '--$this->left' instead.">$this->left -= 1</weak_warning>;
        <weak_warning descr="Use '++$this->left' instead.">$this->left += 1</weak_warning>;
        <weak_warning descr="Use '--$this->start' instead.">$this->start -= 1</weak_warning>;
        <weak_warning descr="Use '--$this->none' instead.">$this->none = $this->none - 1</weak_warning>;
        <weak_warning descr="Use '--$this->zero' instead.">$this->zero -= 1</weak_warning>;
        <weak_warning descr="Use '--self::$depth' instead.">self::$depth -= 1</weak_warning>;
        <weak_warning descr="Use '--static::$depth' instead.">static::$depth -= 1</weak_warning>;
        <weak_warning descr="Use '--self::$level' instead.">self::$level -= 1</weak_warning>;
        <weak_warning descr="Use '--Stock::$depth' instead.">Stock::$depth -= 1</weak_warning>;
        <weak_warning descr="Use '--$this->a' instead.">$this->a -= 1</weak_warning>;
        <weak_warning descr="Use '--$this->b' instead.">$this->b -= 1</weak_warning>;
        <weak_warning descr="Use '--$this->missing' instead.">$this->missing -= 1</weak_warning>;
        <weak_warning descr="Use '--$o->left' instead.">$o->left -= 1</weak_warning>;
        <weak_warning descr="Use '--$this->$name' instead.">$this->$name -= 1</weak_warning>;
        <weak_warning descr="Use '--self::$$name' instead.">self::$$name -= 1</weak_warning>;
        <weak_warning descr="Use '--$u' instead.">$u -= 1</weak_warning>;
        $this->typed -= 1;
        $n -= 1;
        $m = null;
        $m = $m - 1;
        @$u -= 1;
        @<weak_warning descr="Use '++$u' instead.">$u += 1</weak_warning>;
        if ($this->typed !== null) {
            <weak_warning descr="Use '--$this->typed' instead.">$this->typed -= 1</weak_warning>;
        }
    }
}
function outside($o) {
    return function () use ($o) { <weak_warning descr="Use '--$this->left' instead.">$this->left -= 1</weak_warning>; };
}
