<?php
namespace Refs;

function fill(&$target)
{
    $target = ['a', 'b'];
}

function read($value)
{
    return $value;
}

function spread($first, &...$rest)
{
}

function plain($only)
{
}

class Sink
{
    public function __construct(&$slot)
    {
        $slot = 'sunk';
    }

    public function <weak_warning descr="Declare ': void' as the return type.">into</weak_warning>(&...$slots)
    {
    }

    public function <weak_warning descr="Declare ': void' as the return type.">keep</weak_warning>($value, $other = null)
    {
    }

    public static function <weak_warning descr="Declare ': void' as the return type.">swap</weak_warning>(&$value)
    {
        $value = 'swapped';
    }
}

class Plain
{
}

class Orphan extends Missing
{
}

final class Runner
{
    /** @var Sink */
    private $sink;

    private $unknown;

    public function filled()
    {
        $r = '';
        fill($r);
        return $r;
    }

    public function filledNamed()
    {
        $r = '';
        fill(target: $r);
        return $r;
    }

    public function filledElement()
    {
        $r = null;
        fill($r['key']['sub']);
        return $r;
    }

    public function <weak_warning descr="Declare ': string' as the return type.">byValue</weak_warning>()
    {
        $r = '';
        read($r);
        read(value: $r);
        read(other: $r);
        read(...[$r]);
        plain($r, $r);
        $callable = read(...);
        return $r;
    }

    public function variadic()
    {
        $a = 1;
        $b = 2;
        $c = 3;
        spread($a, $b, $c);
        return $c;
    }

    public function <weak_warning descr="Declare ': int' as the return type.">variadicHead</weak_warning>()
    {
        $a = 1;
        $b = 2;
        spread($a, $b);
        return $a;
    }

    public function typed()
    {
        $v = 1;
        settype($v, 'string');
        return $v;
    }

    public function <weak_warning descr="Declare ': array' as the return type.">sorted</weak_warning>()
    {
        $v = [3, 1];
        sort($v);
        return $v;
    }

    public function <weak_warning descr="Declare ': array' as the return type.">matched</weak_warning>(string $s)
    {
        $m = 0;
        preg_match('/\d+/', $s, $m);
        return $m;
    }

    public function alias()
    {
        $a = 1;
        $b = &$a;
        $b = 'text';
        return $a;
    }

    public function aliased()
    {
        $a = 1;
        $b = &$a;
        return $b;
    }

    public function inArray()
    {
        $x = 1;
        $holder = [&$x];
        $holder[0] = 'text';
        return $x;
    }

    public function inList(array $pairs)
    {
        $x = 1;
        [&$x] = $pairs;
        return $x;
    }

    public function looped(array $data)
    {
        $first = 0;
        foreach ($data as &$first) {
        }
        return $first;
    }

    public function <weak_warning descr="Declare ': int' as the return type.">loopedByValue</weak_warning>(array $data)
    {
        $count = 0;
        foreach ($data as $item) {
            $last = $item;
        }
        return $count;
    }

    public function captured()
    {
        $x = 1;
        $set = function () use (&$x) {
            $x = 'text';
        };
        $set();
        return $x;
    }

    public function <weak_warning descr="Declare ': int' as the return type.">capturedByValue</weak_warning>()
    {
        $x = 1;
        $get = function () use ($x) {
            $y = 2;
            fill($y);
            return $x;
        };
        $arrow = fn() => fill($x);
        return $x;
    }

    public function fromGlobal()
    {
        $x = 1;
        global $x;
        return $x;
    }

    public function fromStatic()
    {
        static $n = 1;
        if ($n === 1) {
            $n = 'later';
            return 2;
        }
        return $n;
    }

    public function staticRef()
    {
        $x = 1;
        self::swap($x);
        return $x;
    }

    public function sinkRef()
    {
        $x = 1;
        Sink::swap($x);
        return $x;
    }

    public function methodRef()
    {
        $x = 1;
        $y = 2;
        $this->sink->into($x, $y);
        return $y;
    }

    public function <weak_warning descr="Declare ': int' as the return type.">methodValue</weak_warning>()
    {
        $x = 1;
        $this->sink->keep($x, $x);
        return $x;
    }

    public function constructed()
    {
        $x = 1;
        new Sink($x);
        return $x;
    }

    public function <weak_warning descr="Declare ': int' as the return type.">constructedPlain</weak_warning>()
    {
        $x = 1;
        new Plain($x);
        new Plain;
        return $x;
    }

    public function constructedUnknown()
    {
        $x = 1;
        new Missing($x);
        return $x;
    }

    public function constructedOrphan()
    {
        $x = 1;
        new Orphan($x);
        return $x;
    }

    public function constructedDynamic(string $class)
    {
        $x = 1;
        new $class($x);
        return $x;
    }

    public function dynamicCallee(callable $f)
    {
        $x = 1;
        $f($x);
        return $x;
    }

    public function unknownFunction()
    {
        $x = 1;
        \Elsewhere\touch($x);
        return $x;
    }

    public function unknownReceiver()
    {
        $x = [];
        $this->unknown->load($x);
        return $x;
    }

    public function unknownStatic()
    {
        $x = 1;
        Missing::load($x);
        return $x;
    }

    public function legacyByRef()
    {
        $x = 1;
        $y = &$this->unknown;
        return $y;
    }

    // An object handed to an unknown callee stays that object.
    public function <weak_warning descr="Declare ': Plain' as the return type.">objectToUnknown</weak_warning>()
    {
        $plain = new Plain();
        $this->unknown->register($plain);
        return $plain;
    }

    public function <weak_warning descr="Declare ': ?Plain' as the return type.">nullableObjectToUnknown</weak_warning>(bool $flag)
    {
        $plain = $flag ? new Plain() : null;
        $this->unknown->register($plain);
        return $plain;
    }

    public function scalarOrObjectToUnknown(bool $flag)
    {
        $plain = $flag ? new Plain() : 1;
        $this->unknown->register($plain);
        return $plain;
    }

    public function objectToReference()
    {
        $plain = new Plain();
        fill($plain);
        return $plain;
    }

    // The replaced value reaches the result through these operands.
    public function viaParens()
    {
        $x = 1;
        fill($x);
        return ($x);
    }

    public function viaTernary(bool $flag)
    {
        $x = 1;
        fill($x);
        return $flag ? $x : 2;
    }

    public function viaElse(bool $flag)
    {
        $x = 1;
        fill($x);
        return $flag ? 2 : $x;
    }

    public function viaShortTernary()
    {
        $x = 1;
        fill($x);
        return $x ?: 2;
    }

    public function viaCoalesce()
    {
        $x = 1;
        fill($x);
        return $x ?? 2;
    }

    public function viaArithmetic()
    {
        $x = 1;
        fill($x);
        return 2 + $x;
    }

    public function viaNegation()
    {
        $x = 1;
        fill($x);
        return -$x;
    }

    public function viaAssignment()
    {
        $x = 1;
        fill($x);
        return $y = $x;
    }

    public function viaElement()
    {
        $x = ['a'];
        fill($x);
        return $x[0];
    }

    public function viaMatch(int $k)
    {
        $x = 1;
        fill($x);
        return match ($k) {
            1 => $x,
            default => 2,
        };
    }

    // The result does not take its type from the variable here.
    public function <weak_warning descr="Declare ': string' as the return type.">viaConcat</weak_warning>()
    {
        $x = 1;
        fill($x);
        return 'n' . $x;
    }

    public function <weak_warning descr="Declare ': bool' as the return type.">viaNot</weak_warning>()
    {
        $x = 1;
        fill($x);
        return !$x;
    }

    public function <weak_warning descr="Declare ': int' as the return type.">viaCompound</weak_warning>()
    {
        $x = 1;
        fill($x);
        $y = 1;
        return $y += 2;
    }

    /** @return string */
    public function <weak_warning descr="Declare ': string' as the return type.">documented</weak_warning>()
    {
        $r = '';
        fill($r);
        return $r;
    }
}
