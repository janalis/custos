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
            <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$this->mode == 'on'</warning>,
            <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$this->documented == 'on'</warning>,
            <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$this->typed == 'on'</warning>,
            <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$this->doc() == 'on'</warning>,
            <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$this->native() == 'on'</warning>,
            <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$this->virtual == 'on'</warning>,
            <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$this->described() == 'on'</warning>,
            <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">self::label(1) == 'on'</warning>,
            <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">__CLASS__ == 'Settings'</warning>,
            is_string($this->flag) && <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$this->flag == 'on'</warning>,
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
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$flag == 'yes'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$ref == 'yes'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$rest == 'yes'</warning>,
    ];
}

function expressions($x, string $s, ?string $n, Settings $o, string &$ref = '') {
    $lower = strtolower($x);
    $call = function (): string { return 'a'; };
    return [
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">strtolower($x) == 'on'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">(string) $x == 'on'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">(int) $x == 'on'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">(float) $x == 'on'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">(array) $x == 'on'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">-$x == 'on'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$x . '' == 'on'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">"a$x" == 'on'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$lower == 'on'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">($s ?: 'b') == 'on'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">($x ? $s : 'b') == 'on'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">($n ?? $s) == 'on'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">($n ?? $x) == 'on'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$o->typed == 'on'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$o->native() == 'on'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">max('a', 'b') == 'on'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$call() == 'on'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">Settings::label($x) == 'on'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">Missing::label($x) == 'on'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$o->up()->typed == 'on'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$o->parent?->typed == 'on'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">PHP_OS == 'Linux'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">($s ?: null) == 'on'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">(($s) ?: 'b') == 'on'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">($x ? $s : unknownFunction()) == 'on'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">($x ? unknownFunction() : $s) == 'on'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">(unknownFunction() ?? $s) == 'on'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">match ($x) { 1 => 'a', default => $s } == 'on'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">match ($x) { 1 => 'a', default => unknownFunction() } == 'on'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">match ($x) { 1 => $ref, default => 'a' } == 'on'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$o->$s == 'on'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$o->$s() == 'on'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">['k' => 'a']['k'] == 'on'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$_SERVER['REQUEST_METHOD'] == 'POST'</warning>,
    ];
}

function backEdge(array $xs) {
    $s = 'a';
    foreach ($xs as $x) {
        if (<warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$s == 'yes'</warning>) {
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
    return [<warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$s == 'yes'</warning>, <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$r == 'yes'</warning>];
}

function elementReference() {
    $s = ['k' => 'a'];
    $r = &$s['k'];
    return <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$s == 'yes'</warning>;
}

function itemReference() {
    $s = 'a';
    $refs = [&$s];
    return <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$s == 'yes'</warning>;
}

function loopReference() {
    $list = ['a'];
    foreach ($list as &$item) {
        $item = true;
    }
    return <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$list == 'yes'</warning>;
}

function propertyReference(Settings $o) {
    $r = &$o->typed;
    return <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$o->typed == 'yes'</warning>;
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
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$s == 'yes'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$m == 'yes'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$named == 'yes'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$kept == 'yes'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$spread == 'yes'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$unknown == 'yes'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$dynamic == 'yes'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$partial == 'yes'</warning>,
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
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$a == 'yes'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$b == 'yes'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$c == 'yes'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$d == 'yes'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$e == 'yes'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$f == 'yes'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$g == 'yes'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$h == 'yes'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$i == 'yes'</warning>,
    ];
}

function globals() {
    global $g;
    static $kept = 'a';
    $g = 'a';
    change();
    return [<warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$g == 'yes'</warning>, <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$kept == 'yes'</warning>];
}

function otherWrites(array $pair, string $typed) {
    [$first, $second] = $pair;
    $count = 'a';
    $count++;
    $sum = 'a';
    $early = <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$sum == 'yes'</warning>;
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
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$first == 'yes'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$count == 'yes'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$sum == 'yes'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$text == 'yes'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$late == 'yes'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$gone == 'yes'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$typed == 'yes'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$caught == 'yes'</warning>,
    ];
}

function cycle(string $seed) {
    $a = $seed;
    $b = $a;
    $a = $b;
    return <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$a == 'yes'</warning>;
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
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$one->typed == 'on'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$two->typed == 'on'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$p->typed == 'on'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$shape['k'] == 'on'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$written['k'] == 'on'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$written == 'on'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$text == 'on'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$copy == 'on'</warning>,
    ];
}

function closures(string $outer, string $shared) {
    $early = <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$shared == 'yes'</warning>;
    $byValue = function (string $own) use ($outer) {
        $local = 'a';
        return [<warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$own == 'yes'</warning>, <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$outer == 'yes'</warning>, <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$local == 'yes'</warning>];
    };
    $reassigned = function () use ($outer) {
        $outer = 'a';
        return <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$outer == 'yes'</warning>;
    };
    $byReference = function () use (&$shared) {
        $shared = true;
    };
    $arrow = fn(string $own) => [<warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$own == 'yes'</warning>, <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$outer == 'yes'</warning>];
    return [<warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$outer == 'yes'</warning>, <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$shared == 'yes'</warning>];
}

function variableVariables(string $s, $name) {
    $$name = true;
    return <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$s == 'yes'</warning>;
}

function dynamicName(string $s, $name) {
    return [<warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$$name == 'yes'</warning>, <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">($s ?: $$name) == 'yes'</warning>];
}

function extracted(string $s, array $data) {
    $early = <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$s == 'yes'</warning>;
    extract($data);
    return $early;
}

function included(string $s) {
    include 'vars.php';
    return <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$s == 'yes'</warning>;
}

function evaluated(string $s) {
    eval('$s = true;');
    return <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$s == 'yes'</warning>;
}

function nested(string $s) {
    function inner(&$v) {
        $$v = 1;
    }
    return <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$s == 'yes'</warning>;
}

$top = 'a';
<warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$top == 'yes'</warning>;
// A guard narrows the property, but its declaration still allows a bool.
final class Flagged { public string|bool $flag = 'off'; }
function narrowedFlag(Flagged $o) {
    return is_string($o->flag) && <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$o->flag == 'on'</warning>;
}

function evaluatedLater(string $s) {
    $early = <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$s == 'yes'</warning>;
    eval('$s = true;');
    return $early;
}

function boxedLater(string $s) {
    $early = <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$s == 'yes'</warning>;
    $box = [&$s];
    return $early;
}
