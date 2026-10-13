<?php
class Uploads
{
    private $webRoot;

    public function __construct(string $root)
    {
        $this->webRoot = <warning descr="Use 'dirname($root) . '/public'' instead: realpath() fails inside stream wrappers.">realpath($root . '/../public')</warning>;
        if (!$this->webRoot) {
            throw new \RuntimeException('missing public dir');
        }
    }
}

function a(string $d) {
    if (!<warning descr="Use 'dirname(__DIR__) . '/x'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../x')</warning>) {
        return null;
    }
    if (($p = <warning descr="Use 'dirname(__DIR__) . '/y'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../y')</warning>) === false) {
        return null;
    }
    $q = <warning descr="Use 'dirname(__DIR__) . '/z'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../z')</warning> ?: '/fallback';
    $r = (bool) <warning descr="Use 'dirname(__DIR__) . '/w'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../w')</warning>;
    $s = <warning descr="Use 'dirname(__DIR__) . '/v'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../v')</warning>;
    $t = $s ? 1 : 0;
    $u = <warning descr="Use 'dirname(__DIR__) . '/u'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../u')</warning> && $q;
    while (<warning descr="Use 'dirname(__DIR__) . '/t'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../t')</warning>) {
        break;
    }
    if ($d) {
    } elseif (<warning descr="Use 'dirname(__DIR__) . '/s'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../s')</warning>) {
    }
    $kept = <warning descr="Use 'dirname(__DIR__) . '/k'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../k')</warning>;
    $other = $kept;
    $kept = 'reset';
    $in = [<warning descr="Use 'dirname(__DIR__) . '/i'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../i')</warning>];
    $x = 1 + (int) ($y = <warning descr="Use 'dirname(__DIR__) . '/j'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../j')</warning>);
    $acc = 'x';
    $acc .= <warning descr="Use 'dirname(__DIR__) . '/c'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../c')</warning>;
    if ($d) $single = <warning descr="Use 'dirname(__DIR__) . '/b'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../b')</warning>;
    if (<warning descr="Use 'dirname(__DIR__) . '/a'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../a')</warning>) {
        $acc .= '!';
    }
    return [$p, $q, $r, $t, $u, $other, $in, $x, $acc, $single ?? null];
}
