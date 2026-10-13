<?php
class Vault {
    public int $count;
    public static int $hits;
    public static $shared;
    public ?int $limit = null;
    public $plain;
    public $dropped;
    private $secret;
    protected $guarded;

    public function probe(Vault $peer, ?Vault $maybe, Vault|Crate $either, $flag) {
        $a = isset($this->count);
        $b = isset(self::$hits);
        $c = <weak_warning descr="Compare with null instead: '$this->limit !== null'.">isset($this->limit)</weak_warning>;
        $d = <weak_warning descr="Compare with null instead: '$peer->plain !== null'.">isset($peer->plain)</weak_warning>;
        $e = <weak_warning descr="Compare with null instead: '$maybe->plain !== null'.">isset($maybe->plain)</weak_warning>;
        $f = <weak_warning descr="Compare with null instead: '$either->plain !== null'.">isset($either->plain)</weak_warning>;
        $g = <weak_warning descr="Compare with null instead: '$peer::$shared !== null'.">isset($peer::$shared)</weak_warning>;
        $h = <weak_warning descr="Compare with null instead: '$peer->secret !== null'.">isset($peer->secret)</weak_warning>;
        if ($flag) {
            $late = new Vault();
        }
        $i = <weak_warning descr="Compare with null instead: '$late->plain !== null'.">isset($late->plain)</weak_warning>;
        unset($this->dropped);
        $j = <weak_warning descr="Compare with null instead: '$this->dropped !== null'.">isset($this->dropped)</weak_warning>;
        $k = <weak_warning descr="Compare with null instead: '$this !== null'.">isset($this)</weak_warning>;
        return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $k];
    }
}

class Crate {
}

class SubVault extends Vault {
    public function peek() {
        return [<weak_warning descr="Compare with null instead: '$this->guarded !== null'.">isset($this->guarded)</weak_warning>, isset($this->secret)];
    }
}

class Lazy {
    public $slot;
    public function __get($name) {
        return null;
    }
    public function filled() {
        return <weak_warning descr="Compare with null instead: '$this->slot !== null'.">isset($this->slot)</weak_warning>;
    }
}

trait Stamped {
    private $stamp;
}

class Parcel {
    use Stamped;
    public function stamped() {
        return <weak_warning descr="Compare with null instead: '$this->stamp !== null'.">isset($this->stamp)</weak_warning>;
    }
}

function outside(Vault $v) {
    return [<weak_warning descr="Compare with null instead: '$v->plain !== null'.">isset($v->plain)</weak_warning>, isset($v->secret), isset($v->guarded)];
}

function vars($param, $flag) {
    if ($flag) {
        $cfg = 1;
    }
    $a = <weak_warning descr="Compare with null instead: '$cfg === null'.">!isset($cfg)</weak_warning>;
    static $cache;
    global $shared;
    $b = <weak_warning descr="Compare with null instead: '$cache !== null'.">isset($cache)</weak_warning> && <weak_warning descr="Compare with null instead: '$shared !== null'.">isset($shared)</weak_warning>;
    $c = <weak_warning descr="Compare with null instead: '$$flag !== null'.">isset($$flag)</weak_warning>;
    $gone = 1;
    unset($gone);
    $d = <weak_warning descr="Compare with null instead: '$gone !== null'.">isset($gone)</weak_warning>;
    ($set = 2);
    $e = <weak_warning descr="Compare with null instead: '$set !== null'.">isset($set)</weak_warning>;
    $f = ($inner = 3);
    $g = <weak_warning descr="Compare with null instead: '$inner !== null'.">isset($inner)</weak_warning>;
    switch ($flag) {
        case 1:
            $key = 1;
            $h = <weak_warning descr="Compare with null instead: '$key !== null'.">isset($key)</weak_warning>;
    }
    if ($flag) $loose = 1;
    $i = <weak_warning descr="Compare with null instead: '$loose !== null'.">isset($loose)</weak_warning>;
    $j = <weak_warning descr="Compare with null instead: '$after !== null'.">isset($after)</weak_warning>;
    $after = 1;
    $fn = function ($p) use ($param, &$ref) {
        return [<weak_warning descr="Compare with null instead: '$p !== null'.">isset($p)</weak_warning>, <weak_warning descr="Compare with null instead: '$param !== null'.">isset($param)</weak_warning>, <weak_warning descr="Compare with null instead: '$ref !== null'.">isset($ref)</weak_warning>, <weak_warning descr="Compare with null instead: '$other !== null'.">isset($other)</weak_warning>];
    };
    $arrow = fn($q) => <weak_warning descr="Compare with null instead: '$q !== null'.">isset($q)</weak_warning> && <weak_warning descr="Compare with null instead: '$param !== null'.">isset($param)</weak_warning>;
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $fn, $arrow];
}

function prec($a, $b) {
    return [
        <weak_warning descr="Compare with null instead: '$a !== null'.">isset($a)</weak_warning> === <weak_warning descr="Compare with null instead: '$b !== null'.">isset($b)</weak_warning>,
        (int) <weak_warning descr="Compare with null instead: '$a !== null'.">isset($a)</weak_warning>,
        <weak_warning descr="Compare with null instead: '$a !== null'.">isset($a)</weak_warning> . 'x',
        <weak_warning descr="Compare with null instead: '$a !== null'.">isset($a)</weak_warning> + 1,
        -<weak_warning descr="Compare with null instead: '$a !== null'.">isset($a)</weak_warning>,
        !<weak_warning descr="Compare with null instead: '$a !== null'.">isset($a)</weak_warning> instanceof stdClass,
        !<weak_warning descr="Compare with null instead: '$a === null'.">!isset($a)</weak_warning>,
        <weak_warning descr="Compare with null instead: '$a === null'.">!isset($a)</weak_warning> + 1,
        <weak_warning descr="Compare with null instead: '$a !== null'.">isset($a)</weak_warning> && $b,
        <weak_warning descr="Compare with null instead: '$a !== null'.">isset($a)</weak_warning> ?? $b,
        @<weak_warning descr="Compare with null instead: '$a !== null'.">isset($a)</weak_warning>,
    ];
}

$top = new Vault();
$z = <weak_warning descr="Compare with null instead: '$top->plain !== null'.">isset($top->plain)</weak_warning>;
