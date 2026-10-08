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
    if (!<warning descr="Use 'dirname($d) . '/x'' instead: realpath() fails inside stream wrappers.">realpath($d . '/../x')</warning>) {
        return null;
    }
    if (($p = <warning descr="Use 'dirname($d) . '/y'' instead: realpath() fails inside stream wrappers.">realpath($d . '/../y')</warning>) === false) {
        return null;
    }
    $q = <warning descr="Use 'dirname($d) . '/z'' instead: realpath() fails inside stream wrappers.">realpath($d . '/../z')</warning> ?: '/fallback';
    $r = (bool) <warning descr="Use 'dirname($d) . '/w'' instead: realpath() fails inside stream wrappers.">realpath($d . '/../w')</warning>;
    $s = <warning descr="Use 'dirname($d) . '/v'' instead: realpath() fails inside stream wrappers.">realpath($d . '/../v')</warning>;
    $t = $s ? 1 : 0;
    $u = <warning descr="Use 'dirname($d) . '/u'' instead: realpath() fails inside stream wrappers.">realpath($d . '/../u')</warning> && $q;
    while (<warning descr="Use 'dirname($d) . '/t'' instead: realpath() fails inside stream wrappers.">realpath($d . '/../t')</warning>) {
        break;
    }
    if ($d) {
    } elseif (<warning descr="Use 'dirname($d) . '/s'' instead: realpath() fails inside stream wrappers.">realpath($d . '/../s')</warning>) {
    }
    $kept = <warning descr="Use 'dirname($d) . '/k'' instead: realpath() fails inside stream wrappers.">realpath($d . '/../k')</warning>;
    $other = $kept;
    $kept = 'reset';
    $in = [<warning descr="Use 'dirname($d) . '/i'' instead: realpath() fails inside stream wrappers.">realpath($d . '/../i')</warning>];
    $x = 1 + (int) ($y = <warning descr="Use 'dirname($d) . '/j'' instead: realpath() fails inside stream wrappers.">realpath($d . '/../j')</warning>);
    $acc = 'x';
    $acc .= <warning descr="Use 'dirname($d) . '/c'' instead: realpath() fails inside stream wrappers.">realpath($d . '/../c')</warning>;
    if ($d) $single = <warning descr="Use 'dirname($d) . '/b'' instead: realpath() fails inside stream wrappers.">realpath($d . '/../b')</warning>;
    if (<warning descr="Use 'dirname($d) . '/a'' instead: realpath() fails inside stream wrappers.">realpath($d . '/../a')</warning>) {
        $acc .= '!';
    }
    return [$p, $q, $r, $t, $u, $other, $in, $x, $acc, $single ?? null];
}
