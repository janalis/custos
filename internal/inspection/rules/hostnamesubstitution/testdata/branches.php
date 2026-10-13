<?php
$name = 'x';
$to = "$user@" . <error descr="E-mail address built from client-controlled $_SERVER['HTTP_HOST']; validate it against a whitelist.">$_SERVER['HTTP_HOST']</error>;
$plain = $prefix . $_SERVER['HTTP_HOST'];
$$name = $_SERVER['SERVER_NAME'];

function overwrittenInCondition()
{
    $host = $_SERVER['HTTP_HOST'];
    if ($host = 'example.org') {
    }
    return 'a@' . <error descr="E-mail address built from client-controlled $_SERVER['HTTP_HOST']; validate it against a whitelist.">$host</error>;
}

class Config
{
    // Not valid PHP (a default must be a constant expression).
    public $host = $_SERVER['HTTP_HOST'];
}
