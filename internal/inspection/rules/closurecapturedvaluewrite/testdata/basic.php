<?php
$n = 0; $f = <warning descr="Capture the scalar by reference when updating outer state.">function () use ($n) { ++$n; }</warning>; $f(); echo $n;

function grow(string $text): void {
    $pad = <warning descr="Capture the scalar by reference when updating outer state.">function () use ($text) {
        $text = $text . '!';
        ($text .= '?');
    }</warning>;
    $pad();
    echo $text;
}
