<?php
function configure()
{
    global $settingsMode, $registryList;
    $settingsMode = 'compact';
    $registryList[] = 'entry';
}

function tally(array &$stats)
{
    $bucket = &$stats['bucket'];
    $bucket['hits']++;
}

function collect()
{
    $bag = new \ArrayObject();
    $bag[] = 'first';
}

function stillReported()
{
    $list = [];
    <weak_warning descr="Value is only written here and never read; the write is lost.">$list[]</weak_warning> = 2;
    $n = 0;
    <weak_warning descr="Value is only written here and never read; the write is lost.">$n['k']</weak_warning>++;
}

function nestedScopeObjects()
{
    // An object assigned in a nested closure belongs to another scope.
    $make = function () { $pile = new \ArrayObject(); return $pile; };
    $pile = [];
    <weak_warning descr="Value is only written here and never read; the write is lost.">$pile[]</weak_warning> = $make;
}
