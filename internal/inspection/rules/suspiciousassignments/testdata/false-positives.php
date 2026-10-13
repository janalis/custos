<?php
function classify($kind) {
    switch ($kind) {
        case 'x':
            $label = 'ex';
            break;
        case 'y':
            $label = 'why';
        case 'z':
            $label = $label . 'z';
            if ($kind) { $other = 1; }
        case 'w':
            $other = 2;
    }
}

$path .= $path . '/' . 'x';
$total *= ($total * 3);
$total += ($total) + 1;
$total += $other - $total;

function normalize($name, $size, &$out, $flag) {
    $name = strtolower($name);
    $out = [];
    if ($flag) { $size = 1; }
    $flag = 2;
    $cb = function () use ($size) { return $size; };
    return [$name, $size];
}

function lonely($p) {
    $p = 1;
}

function captured($p) {
    $p = function () use ($p) { return $p; };
    return $p;
}

$delta = -$step;
$other=+$step;
$third =-$step;

function flow($cond, $list, $i) {
    if ($cond) { $cfg["k"] = 1; return; }
    $cfg["k"] = 2;
    if ($cond) { $mode = 'a'; } else { $mode = 'c'; }
    $mode = 'b';
    if ($cond) { $x = 1; echo $x; }
    $x = 2;
    $list[] = 1;
    $list[] = 2;
    $tmp = 'x';
    $tmp = "<{$tmp}>";
    $v = trim($v);
    $v = trim($v) . $v;
    $a[$i] = 1;
    $a[$i] = 2;
    $b[__LINE__] = 1;
    $b[__LINE__] = 2;
    $r = &$list;
    $r = 3;
    $s .= 'x';
    $s = '';
    try {
        $t = 1;
        $t = 2;
    } catch (Exception $e) {}
    $m[count($m)] = 1;
    $m[count($m)] = 2;
    return [$mode, $cfg, $tmp];
}

function unpack_all(array $row, \ArrayObject $obj, mixed $m, ?array $maybe) {
    [$a, $b] = $row;
    list($c, $d) = $obj;
    [$e] = $m;
    [$f] = $maybe;
    [$g] = $unknown;
    foreach ($row as [$h, $i]) {}
}

class NormalizeTest {
    public function run($p) {
        $p = 1;
        return $p;
    }
}

function compound_in_if($c) {
    if ($c) { $acc .= 'x'; }
    $acc = '';
    $same = 1;
    $same = $same;
    return [$acc, $same];
}
