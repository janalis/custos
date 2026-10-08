<?php
class Uploads
{
    private $webRoot;

    public function __construct(string $root)
    {
        $this->webRoot = realpath($root . '/../public');
        if (!$this->webRoot) {
            throw new \RuntimeException('missing public dir');
        }
    }
}

function a(string $d) {
    if (!realpath($d . '/../x')) {
        return null;
    }
    if (($p = realpath($d . '/../y')) === false) {
        return null;
    }
    $q = realpath($d . '/../z') ?: '/fallback';
    $r = (bool) realpath($d . '/../w');
    $s = realpath($d . '/../v');
    $t = $s ? 1 : 0;
    $u = realpath($d . '/../u') && $q;
    while (realpath($d . '/../t')) {
        break;
    }
    if ($d) {
    } elseif (realpath($d . '/../s')) {
    }
    $kept = dirname($d) . '/k';
    $other = $kept;
    $kept = 'reset';
    $in = [dirname($d) . '/i'];
    $x = 1 + (int) ($y = dirname($d) . '/j');
    $acc = 'x';
    $acc .= dirname($d) . '/c';
    if ($d) $single = dirname($d) . '/b';
    if (realpath($d . '/../a')) {
        $acc .= '!';
    }
    return [$p, $q, $r, $t, $u, $other, $in, $x, $acc, $single ?? null];
}
