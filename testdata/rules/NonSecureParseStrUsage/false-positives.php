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
