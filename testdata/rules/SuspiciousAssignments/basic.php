<?php
function classify($kind) {
    switch ($kind) {
        case 'x':
            $label = 'ex';
        case 'y':
            <error descr="This write overwrites a value set in a previous case; a 'break' may be missing.">$label</error> = 'why';
            break;
        case 'p':
            $pair = 1;
        case 'q':
            [<error descr="This write overwrites a value set in a previous case; a 'break' may be missing.">$pair</error>, $rest] = [2, 3];
            return $pair;
        case 'r':
        case 's':
            $acc[] = $kind;
            $msg = 'm';
            $msg = strtoupper($msg);
            break;
    }
}

<error descr="The target is repeated on the right-hand side of the compound assignment; likely a merge mistake.">$total -= $total - 1</error>;
<error descr="The target is repeated on the right-hand side of the compound assignment; likely a merge mistake.">$path .= $path . '/'</error>;

function normalize($name, $size, &$out) {
    $name = strtolower($name);
    <error descr="Parameter is overwritten before its value is used.">$size</error> = strlen($name);
    $out = [];
    return [$name, $size];
}

<error descr="Did you mean '-='? Fix the operator or the spacing.">$delta =- $step</error>;
<error descr="Did you mean '!='? Fix the operator or the spacing.">$ok =! $failed</error>;

function flow($cond, $list) {
    if ($cond) { $mode = 'a'; }
    <error descr="$mode is overwritten right after the 'if'; an 'else' may be missing.">$mode = 'b'</error>;

    $count = 0;
    <error descr="$count is overwritten right after being assigned.">$count</error> = count($list);

    $map['a'][1] = 1;
    <error descr="$map['a'][1] is overwritten right after being assigned.">$map['a'][1]</error> = 2;
    return [$mode, $count, $map];
}

function unpack_all(int $n, array $row, \ArrayObject $obj) {
    <error descr="Destructuring a value that is not an array.">[$e, $f] = $n</error>;
    <error descr="Destructuring a value that is not an array.">list($g) = new \DateTime()</error>;
}
