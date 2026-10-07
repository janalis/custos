<?php
$top = compact('nothing');

function flow($flag)
{
    if ($flag) {
        $maybe = 1;
    }
    global $config;
    static $cache;
    echo $read;
    return compact('maybe', 'config', 'cache', 'read', 'flag', '', "x$flag", $flag . 'y');
}

function nested()
{
    $cb = function ($inner) { return $inner; };
    return compact('inner', 'cb');
}

function noArgs()
{
    return compact();
}

function method()
{
    return $obj->compact('missing');
}
