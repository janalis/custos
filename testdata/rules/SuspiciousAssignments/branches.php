<?php
function cases($k, $data) {
    switch ($k) {
        case 1:
            $a = 1;
            goto done;
        case 2:
            $a = 2;
            exit;
        case 3:
            $a = 3;
            throw new Exception('x');
        case 4:
            $a = 4;
            $n += 1;
            list($b, $c) = $data;
            list(, $d) = $data;
            [[$e], $f] = $data;
        case 5:
            list(<error descr="This write overwrites a value set in a previous case; a 'break' may be missing.">$b</error>) = $data;
            break;
    }
    done:
    return 0;
}

function flows($p, $q, $r) {
    ${$p} = 1;
    $g = fn () => $q;
    function inner($q) { return $q; }
    <error descr="Parameter is overwritten before its value is used.">$r</error> = 1;
    return $r;
}

function others() {
    $a = ~$b;
    $c = @$d;
    $x = $y = 1;
    <error descr="$x is overwritten right after being assigned.">$x</error> = $y = 2;
    $arr[-1] = 1;
    <error descr="$arr[-1] is overwritten right after being assigned.">$arr[-1]</error> = 2;
    $arr[-'k'] = 1;
    $arr[-'k'] = 2;
    $arr[+1] = 1;
    $arr[+1] = 2;
    $o->{$o} = 1;
    $o->{$o} = 2;
    $o->{$i++} = 1;
    $o->{$i++} = 2;
    if ($cond) {
        $flag = 1;
        if ($flag) {
            echo 1;
        }
    }
    $flag = 2;
    if ([$m, $n] = $pair) {}
}

function access(ArrayAccess $x, int $i) {
    [$a, $b] = $x;
    <error descr="Destructuring a value that is not an array.">[$c] = $i</error>;
}
