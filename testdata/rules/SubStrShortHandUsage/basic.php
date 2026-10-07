<?php
function cut($name, $head, $k) {
    $a = substr($name, 0, <warning descr="Pass '-4' as the length instead.">strlen($name) - 4</warning>);
    $b = mb_substr($name, 3, <warning descr="Pass '-2' as the length instead.">mb_strlen($name) - 5</warning>, 'UTF-8');
    $c = \substr($name, 7, <warning descr="The length 'strlen($name) - 2' is unnecessary; remove it.">strlen($name) - 2</warning>);
    $d = mb_substr($name, $k, <warning descr="The length 'mb_strlen($name) - $k' is unnecessary; remove it.">mb_strlen($name) - $k</warning>);
    $e = mb_substr(
        $name,
        strlen($head),
        <warning descr="The length 'mb_strlen($name) - strlen($head)' is unnecessary; remove it.">mb_strlen($name) - strlen($head)</warning>,
        'UTF-8'
    );
    $f = substr($name, 010, <warning descr="Pass '-2' as the length instead.">strlen( $name ) - 12</warning>);
    $g = substr($this->v, 1, <warning descr="The length 'strlen($this->v) - 1' is unnecessary; remove it.">strlen($this->v) - 1</warning>);
    return [$a, $b, $c, $d, $e, $f, $g];
}
