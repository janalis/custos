<?php
function cloneAll(array $values)
{
    // Holds the objects so that their ids are not reused while cloning.
    $objects = [];
    $refs = [];
    $out = [];
    foreach ($values as $k => $v) {
        if (!is_object($v)) {
            continue;
        }
        $h = spl_object_id($v);
        if (empty($refs[$h])) {
            $refs[$h] = $k;
            $objects[] = $v;
        }
        $out[$k] = $h;
    }
    return $out;
}

function hashes(array $values, \stdClass $o)
{
    $held = [];
    $held[] = $o;
    $names = [];
    foreach ($values as $v) {
        <weak_warning descr="Value is only written here and never read; the write is lost.">$names[]</weak_warning> = (string) $v;
    }
    return spl_object_hash($o);
}

function noIdentity(\stdClass $o)
{
    $held = [];
    <weak_warning descr="Value is only written here and never read; the write is lost.">$held[]</weak_warning> = $o;
    $f = function ($x) { return spl_object_id($x); };
    return $f;
}
