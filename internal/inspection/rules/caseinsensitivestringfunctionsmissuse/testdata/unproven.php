<?php
// The needle variable can hold something else than the literal: no report.
if (PHP_OS_FAMILY === 'Windows') {
    define('DS', '\\');
} else {
    define('DS', '/');
}
define($dynamicName, 'b');

class Holder
{
    private $sep = '-';

    public function cut($text)
    {
        return stripos($text, $this->sep);
    }
}

function with_default($path, $sep = '/')
{
    return stripos($path, $sep);
}

function reassigned($path, $other, $flag)
{
    $n = '/';
    if ($flag) {
        $n = $other;
    }
    return stripos($path, $n);
}

function looped($path, array $seps)
{
    $n = '-';
    foreach ($seps as $n) {
        if (stripos($path, $n) !== false) {
            return $n;
        }
    }
    return null;
}

function named($path)
{
    return stripos(needle: 'ABC', haystack: '/');
}

function chained($path)
{
    $m = '#';
    $n = $m;
    return stripos($path, $n);
}

function extracted($path, array $vars)
{
    $n = '#';
    extract($vars);
    return stripos($path, $n);
}

function variable_variable($path, $key, $value)
{
    $n = '#';
    $$key = $value;
    return stripos($path, $n);
}

function included($path)
{
    $n = '#';
    include 'overrides.php';
    return stripos($path, $n);
}

function evaluated($path, $code)
{
    $n = '#';
    eval($code);
    return stripos($path, $n);
}

function aliased($path)
{
    $n = '#';
    $alias = &$n;
    $alias = 'b';
    return stripos($path, $n);
}

function array_aliased($path)
{
    $n = '#';
    $box = [&$n];
    $box[0] = 'b';
    return stripos($path, $n);
}

function &generated($path)
{
    $n = '#';
    yield $n;
    return stripos($path, $n);
}

function out_parameter($path)
{
    $n = '#';
    preg_match('/b/i', $path, $n);
    return stripos($path, $n);
}

function unknown_callee($path)
{
    $n = '#';
    mystery_helper($n);
    return stripos($path, $n);
}

function conditional_constant($path)
{
    return stripos($path, DS);
}

function expression($path, $suffix)
{
    return stripos($path, '/' . $suffix);
}
