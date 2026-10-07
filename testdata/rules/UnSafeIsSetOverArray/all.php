<?php
$grid = [];
if (isset(<warning descr="Compute the concatenated key in a variable before using it.">$grid['r' . $row]</warning>)) {
    $both = isset($grid['a'], $grid['b']);
    $cell = isset($grid['k']) ? $grid['k'] : NULL;
    $cell = !isset($grid['k']) ? null : $grid['k'];
    print isset(<weak_warning descr="Use array_key_exists() to check for the key itself.">$grid[$col]</weak_warning>);
    $seen = !isset(<weak_warning descr="Use array_key_exists() to check for the key itself.">$grid['r' . $row]</weak_warning>);
}
$top = isset($grid);

class Ledger {
    public $total;
    public static $shared;
    public function __construct(private $promoted = null) {}
    public function check(\ArrayAccess $bag, $field, ?Ledger $other) {
        $tmp = new stdClass();
        $u = isset($bag['x']);
        $v = isset($tmp->anything);
        $w = isset($this->$field);
        $x = <weak_warning descr="Compare with null instead: '$this->promoted !== null'.">isset($this->promoted)</weak_warning>;
        $p = <weak_warning descr="Compare with null instead: '$this->total !== null'.">isset($this->total)</weak_warning>;
        $q = <weak_warning descr="Compare with null instead: '$tmp === null'.">!isset($tmp)</weak_warning>;
        $s = <weak_warning descr="Compare with null instead: 'self::$shared !== null'.">isset(self::$shared)</weak_warning>;
        $o = <weak_warning descr="Compare with null instead: '$other->total !== null'.">isset($other->total)</weak_warning>;
        $n = !(<weak_warning descr="Compare with null instead: '$field !== null'.">isset($field)</weak_warning>);
        $z = isset($grid['a']) ? $grid['a'] : \null;
        $y = isset(<weak_warning descr="Use array_key_exists() to check for the key itself.">$grid['a']</weak_warning>) ? $grid['a'] : (null);
        try {
            return 1;
        } finally {
            $r = isset($tmp);
        }
    }
}
