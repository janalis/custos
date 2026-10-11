<?php
// The strict fix needs proof that the other operand never holds a bool or
// an object: native declarations only, on variables nothing else can rebind.
/**
 * @property string $virtual
 * @method string described()
 */
class Settings {
    public $mode = 'off';
    public ?Settings $parent = null;
    /** @var string */
    public $documented = 'off';
    public string $typed = 'off';
    public string|bool $flag = 'off';
    public function __construct(string &$out = '') {}
    /** @return string */
    public function doc() { return $this->mode; }
    public function native(): string { return 'x'; }
    public function own() {
        return [
            $this->mode == 'on',
            $this->documented == 'on',
            $this->typed === 'on',
            $this->doc() == 'on',
            $this->native() === 'on',
            $this->virtual == 'on',
            $this->described() == 'on',
            self::label(1) === 'on',
            __CLASS__ === 'Settings',
            is_string($this->flag) && $this->flag == 'on',
        ];
    }
    public function up(): Settings { return $this; }
    public static function make(string $a, string &$b) {}
    public static function label($v): string { return 'x'; }
    public function __get($name) { return true; }
    public function __call($name, $args) { return true; }
}
class Plain {}
class Orphan extends MissingParent {}

/** @param string $flag */
function documented($flag, string &$ref, string ...$rest) {
    return [
        $flag == 'yes',
        $ref == 'yes',
        $rest === 'yes',
    ];
}

function expressions($x, string $s, ?string $n, Settings $o, string &$ref = '') {
    $lower = strtolower($x);
    $call = function (): string { return 'a'; };
    return [
        strtolower($x) === 'on',
        (string) $x === 'on',
        (int) $x === 'on',
        (float) $x === 'on',
        (array) $x === 'on',
        -$x == 'on',
        $x . '' === 'on',
        "a$x" === 'on',
        $lower === 'on',
        ($s ?: 'b') === 'on',
        ($x ? $s : 'b') === 'on',
        ($n ?? $s) === 'on',
        ($n ?? $x) == 'on',
        $o->typed === 'on',
        $o->native() === 'on',
        max('a', 'b') == 'on',
        $call() == 'on',
        Settings::label($x) === 'on',
        Missing::label($x) == 'on',
        $o->up()->typed === 'on',
        $o->parent?->typed === 'on',
        PHP_OS === 'Linux',
        ($s ?: null) === 'on',
        (($s) ?: 'b') === 'on',
        ($x ? $s : unknownFunction()) == 'on',
        ($x ? unknownFunction() : $s) == 'on',
        (unknownFunction() ?? $s) == 'on',
        match ($x) { 1 => 'a', default => $s } === 'on',
        match ($x) { 1 => 'a', default => unknownFunction() } == 'on',
        match ($x) { 1 => $ref, default => 'a' } == 'on',
        $o->$s == 'on',
        $o->$s() == 'on',
        ['k' => 'a']['k'] == 'on',
        $_SERVER['REQUEST_METHOD'] == 'POST',
    ];
}

function backEdge(array $xs) {
    $s = 'a';
    foreach ($xs as $x) {
        if ($s == 'yes') {
            return 'match';
        }
        $s = $x;
    }
    return 'no match';
}

function reference() {
    $s = 'a';
    $r = &$s;
    $r = true;
    return [$s == 'yes', $r == 'yes'];
}

function elementReference() {
    $s = ['k' => 'a'];
    $r = &$s['k'];
    return $s == 'yes';
}

function itemReference() {
    $s = 'a';
    $refs = [&$s];
    return $s == 'yes';
}

function loopReference() {
    $list = ['a'];
    foreach ($list as &$item) {
        $item = true;
    }
    return $list == 'yes';
}

function propertyReference(Settings $o) {
    $r = &$o->typed;
    return $o->typed == 'yes';
}

function byReferenceArgument(string $in) {
    $s = 'a';
    settype($s, 'bool');
    $m = '';
    sscanf($in, '%d', $in, $m);
    $named = 'a';
    settype(type: 'bool', var: $named);
    $kept = 'a';
    strlen(string: $kept);
    $spread = ['a'];
    strlen(...$spread);
    $unknown = 'a';
    unknownFunction($unknown);
    $dynamic = 'a';
    $in($dynamic);
    $partial = 'a';
    $callable = strlen(...);
    return [
        $s == 'yes',
        $m == 'yes',
        $named == 'yes',
        $kept === 'yes',
        $spread == 'yes',
        $unknown == 'yes',
        $dynamic == 'yes',
        $partial === 'yes',
    ];
}

function methodArguments(Settings $o, $any) {
    $a = 'a';
    $b = 'b';
    $o->native($a);
    Settings::make($a, $b);
    $c = 'c';
    $any->run($c);
    $d = 'd';
    new Settings($d);
    $e = 'e';
    new Plain($e);
    $f = 'f';
    new Orphan($f);
    $g = 'g';
    new $any($g);
    $h = 'h';
    new class($h) {};
    $i = 'i';
    new Plain;
    return [
        $a === 'yes',
        $b == 'yes',
        $c == 'yes',
        $d == 'yes',
        $e === 'yes',
        $f == 'yes',
        $g == 'yes',
        $h == 'yes',
        $i === 'yes',
    ];
}

function globals() {
    global $g;
    static $kept = 'a';
    $g = 'a';
    change();
    return [$g == 'yes', $kept == 'yes'];
}

function otherWrites(array $pair, string $typed) {
    [$first, $second] = $pair;
    $count = 'a';
    $count++;
    $sum = 'a';
    $early = $sum == 'yes';
    $sum += 1;
    $text = 'a';
    $text .= $pair[0];
    $late ??= 'a';
    $late ??= $pair[0];
    $gone = 'a';
    unset($gone);
    $typed = 'b';
    try {
        change();
    } catch (Exception $caught) {
    }
    $caught = 'a';
    return [
        $first == 'yes',
        $count == 'yes',
        $sum == 'yes',
        $text === 'yes',
        $late == 'yes',
        $gone == 'yes',
        $typed === 'yes',
        $caught == 'yes',
    ];
}

function cycle(string $seed) {
    $a = $seed;
    $b = $a;
    $a = $b;
    return $a == 'yes';
}

function bases(Settings $o, Settings $p, $any) {
    $one = new Settings();
    $two = new Settings();
    $two = $any;
    $p = new Settings();
    $shape = ['k' => 'a'];
    $written = ['k' => 'a'];
    $written['k'] = $any;
    $text = 'a';
    $text[0] = 'b';
    $from = $any;
    $copy = $from->settings;
    return [
        $one->typed === 'on',
        $two->typed == 'on',
        $p->typed == 'on',
        $shape['k'] == 'on',
        $written['k'] == 'on',
        $written === 'on',
        $text === 'on',
        $copy == 'on',
    ];
}

function closures(string $outer, string $shared) {
    $early = $shared == 'yes';
    $byValue = function (string $own) use ($outer) {
        $local = 'a';
        return [$own === 'yes', $outer == 'yes', $local === 'yes'];
    };
    $reassigned = function () use ($outer) {
        $outer = 'a';
        return $outer == 'yes';
    };
    $byReference = function () use (&$shared) {
        $shared = true;
    };
    $arrow = fn(string $own) => [$own === 'yes', $outer == 'yes'];
    return [$outer === 'yes', $shared == 'yes'];
}

function variableVariables(string $s, $name) {
    $$name = true;
    return $s == 'yes';
}

function dynamicName(string $s, $name) {
    return [$$name == 'yes', ($s ?: $$name) == 'yes'];
}

function extracted(string $s, array $data) {
    $early = $s == 'yes';
    extract($data);
    return $early;
}

function included(string $s) {
    include 'vars.php';
    return $s == 'yes';
}

function evaluated(string $s) {
    eval('$s = true;');
    return $s == 'yes';
}

function nested(string $s) {
    function inner(&$v) {
        $$v = 1;
    }
    return $s === 'yes';
}

$top = 'a';
$top == 'yes';
