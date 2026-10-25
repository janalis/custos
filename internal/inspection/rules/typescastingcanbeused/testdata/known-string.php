<?php
class Paths { public string $root = '/'; }

function paths(string $dir, ?string $maybe, int $n, Paths $p) {
    $made = sprintf('%s/x', $dir);
    $a = "$dir";
    $b = "$made";
    $c = "{$p->root}";
    $d = <weak_warning descr="Use '(string)$maybe' to make the string conversion explicit.">"$maybe"</weak_warning>;
    $e = <weak_warning descr="Use '(string)$n' to make the string conversion explicit.">"$n"</weak_warning>;
    return [$a, $b, $c, $d, $e];
}
