<?php
function readQuery($raw, $q)
{
    parse_str($raw, $fields);
    mb_parse_str($raw, $more);
    $q->parse_str($raw);
    Q::parse_str($raw);
    parse_str();
    Ns\parse_str($raw);
    return [$fields, $more];
}

function viaClosure($raw)
{
    $parse = parse_str(...);
    $parse($raw, $fields);
    return $fields;
}

function viaSpread($raw)
{
    $fields = [];
    $args = [$raw, &$fields];
    parse_str(...$args);
    \mb_parse_str(...$args);
    return $fields;
}
