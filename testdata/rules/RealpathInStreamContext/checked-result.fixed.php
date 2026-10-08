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
    if (!realpath(__DIR__ . '/../x')) {
        return null;
    }
    if (($p = realpath(__DIR__ . '/../y')) === false) {
        return null;
    }
    $q = realpath(__DIR__ . '/../z') ?: '/fallback';
    $r = (bool) realpath(__DIR__ . '/../w');
    $s = realpath(__DIR__ . '/../v');
    $t = $s ? 1 : 0;
    $u = realpath(__DIR__ . '/../u') && $q;
    while (realpath(__DIR__ . '/../t')) {
        break;
    }
    if ($d) {
    } elseif (realpath(__DIR__ . '/../s')) {
    }
    $kept = dirname(__DIR__) . '/k';
    $other = $kept;
    $kept = 'reset';
    $in = [dirname(__DIR__) . '/i'];
    $x = 1 + (int) ($y = dirname(__DIR__) . '/j');
    $acc = 'x';
    $acc .= dirname(__DIR__) . '/c';
    if ($d) $single = dirname(__DIR__) . '/b';
    if (realpath(__DIR__ . '/../a')) {
        $acc .= '!';
    }
    return [$p, $q, $r, $t, $u, $other, $in, $x, $acc, $single ?? null];
}
