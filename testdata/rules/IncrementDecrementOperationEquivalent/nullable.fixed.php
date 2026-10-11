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
        $this->left -= 1;
        ++$this->left;
        --$this->start;
        $this->none = $this->none - 1;
        --$this->zero;
        self::$depth -= 1;
        static::$depth -= 1;
        --self::$level;
        --Stock::$depth;
        $this->a -= 1;
        --$this->b;
        --$this->missing;
        --$o->left;
        --$this->$name;
        --self::$$name;
        --$u;
        $this->typed -= 1;
        $n -= 1;
        $m = null;
        $m = $m - 1;
        @$u -= 1;
        @++$u;
        if ($this->typed !== null) {
            --$this->typed;
        }
    }
}
function outside($o) {
    return function () use ($o) { --$this->left; };
}
