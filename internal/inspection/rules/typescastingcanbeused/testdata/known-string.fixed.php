<?php
class Paths { public string $root = '/'; }

function paths(string $dir, ?string $maybe, int $n, Paths $p) {
    $made = sprintf('%s/x', $dir);
    $a = "$dir";
    $b = "$made";
    $c = "{$p->root}";
    $d = (string)$maybe;
    $e = (string)$n;
    return [$a, $b, $c, $d, $e];
}
