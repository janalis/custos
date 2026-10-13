<?php
// A variable written without an assignment naming it has no complete value.
function pick(&$out)
{
    $out = 'string';
}

function label($in, &...$rest)
{
    return $in;
}

function byRefArgument($v)
{
    $t = 'array';
    pick($t);
    return gettype($v) === $t;
}

function byRefVariadic($v)
{
    $t = 'array';
    sscanf('string', '%s', $t);
    return gettype($v) === $t;
}

function byRefTail($v)
{
    $t = 'array';
    label(1, 2, $t);
    return gettype($v) === $t;
}

function namedByRef($v)
{
    $t = 'array';
    pick(out: $t);
    return gettype($v) === $t;
}

function unknownName($v)
{
    $t = 'array';
    pick(other: $t);
    return gettype($v) === $t;
}

function callTimeReference($v)
{
    $t = 'array';
    strlen(&$t);
    return gettype($v) === $t;
}

function unresolvedCallee($v, $o, $fn)
{
    $t = 'array';
    $o->fill($t);
    $a = 'array';
    Registry::fill($a);
    $b = 'array';
    new Filler($b);
    $c = 'array';
    $fn($c);
    return [gettype($v) === $t, gettype($v) === $a, gettype($v) === $b, gettype($v) === $c];
}

function alias($v)
{
    $t = 'array';
    $r = &$t;
    $r = 'string';
    return gettype($v) === $t;
}

function arrayAlias($v)
{
    $t = 'array';
    $u = [&$t];
    $u[0] = 'string';
    return gettype($v) === $t;
}

function derived($v)
{
    $s = 'array';
    pick($s);
    $t = $s;
    return gettype($v) === $t;
}

function extracted($v, array $o)
{
    $t = 'array';
    extract($o);
    return gettype($v) === $t;
}

function variableVariable($v, $n)
{
    $t = 'array';
    $$n = 'string';
    return gettype($v) === $t;
}

function included($v)
{
    $t = 'array';
    include 'kinds.php';
    return gettype($v) === $t;
}

function evaluated($v, $code)
{
    $t = 'array';
    eval($code);
    return gettype($v) === $t;
}

// Still complete: the variable is only read.
function byValue($v, $o)
{
    $t = 'array';
    new Plain;
    strlen($t);
    label($t);
    label(in: $t);
    strlen($t, 1, 2);
    $o->fill($v);
    $u = [$t];
    $w = $u;
    $w = $t;
    return <warning descr="Use 'is_array($v)' instead.">gettype($v) === $t</warning>;
}

const KIND = 'object';

function constant_($v)
{
    return <warning descr="Use 'is_object($v)' instead.">gettype($v) === KIND</warning>;
}

$top = <warning descr="Use 'is_object($y)' instead.">gettype($y) === KIND</warning>;
