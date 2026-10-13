<?php
<error descr="Supply a nonempty padding string.">str_pad('q', 8, '')</error>;

function padWithVariables(): string
{
    $width = 9;
    $pad = '';
    return <error descr="Supply a nonempty padding string.">str_pad('ab', $width, $pad)</error>;
}

function padBeforeStep(): string
{
    $width = 9;
    $out = <error descr="Supply a nonempty padding string.">str_pad('ab', $width, '')</error>;
    $width--;
    return $out;
}
