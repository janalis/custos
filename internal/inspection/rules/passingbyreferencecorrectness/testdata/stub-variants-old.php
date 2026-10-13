<?php
function timeout() { return 1.0; }
function code() { return 0; }

function connect($host, $port)
{
    $a = fsockopen($host, $port, $errno, $errstr, timeout());
    $b = fsockopen($host, $port, <warning descr="Pass a variable here: this parameter is taken by reference.">code()</warning>, $errstr, timeout());
    return [$a, $b];
}
