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
        $c = $this->limit !== null;
        $d = $peer->plain !== null;
        $e = isset($maybe->plain);
        $f = isset($either->plain);
        $g = isset($peer::$shared);
        $h = $peer->secret !== null;
        if ($flag) {
            $late = new Vault();
        }
        $i = isset($late->plain);
        unset($this->dropped);
        $j = isset($this->dropped);
        $k = isset($this);
        return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $k];
    }
}

class Crate {
}

class SubVault extends Vault {
    public function peek() {
        return [$this->guarded !== null, isset($this->secret)];
    }
}

class Lazy {
    public $slot;
    public function __get($name) {
        return null;
    }
    public function filled() {
        return isset($this->slot);
    }
}

trait Stamped {
    private $stamp;
}

class Parcel {
    use Stamped;
    public function stamped() {
        return $this->stamp !== null;
    }
}

function outside(Vault $v) {
    return [$v->plain !== null, isset($v->secret), isset($v->guarded)];
}

function vars($param, $flag) {
    if ($flag) {
        $cfg = 1;
    }
    $a = !isset($cfg);
    static $cache;
    global $shared;
    $b = $cache !== null && $shared !== null;
    $c = isset($$flag);
    $gone = 1;
    unset($gone);
    $d = isset($gone);
    ($set = 2);
    $e = $set !== null;
    $f = ($inner = 3);
    $g = isset($inner);
    switch ($flag) {
        case 1:
            $key = 1;
            $h = $key !== null;
    }
    if ($flag) $loose = 1;
    $i = isset($loose);
    $j = isset($after);
    $after = 1;
    $fn = function ($p) use ($param, &$ref) {
        return [$p !== null, $param !== null, $ref !== null, isset($other)];
    };
    $arrow = fn($q) => $q !== null && isset($param);
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $fn, $arrow];
}

function prec($a, $b) {
    return [
        ($a !== null) === ($b !== null),
        (int) ($a !== null),
        ($a !== null) . 'x',
        ($a !== null) + 1,
        -($a !== null),
        !($a !== null) instanceof stdClass,
        !($a === null),
        ($a === null) + 1,
        $a !== null && $b,
        $a !== null ?? $b,
        @($a !== null),
    ];
}

$top = new Vault();
$z = isset($top->plain);
