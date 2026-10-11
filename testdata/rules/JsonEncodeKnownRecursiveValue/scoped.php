<?php

function selfReference(): string
{
    $label = 'cycle';
    $a = [];
    $a['self'] = &$a;
    echo $label;
    return <error descr="Remove the cycle before JSON encoding.">json_encode($a)</error>;
}

function mutualReference(): string
{
    $a = [];
    $b = [];
    $a['b'] = &$b;
    $b['a'] = &$a;
    return <error descr="Remove the cycle before JSON encoding.">json_encode($a)</error>;
}

function assignedResult(): string
{
    $a = [];
    $a['self'] = &$a;
    $json = <error descr="Remove the cycle before JSON encoding.">json_encode($a)</error>;
    return $json;
}
