<?php
function timeout() { return 1.0; }
function code() { return 0; }

// Built-ins whose stubs repeat a parameter per PHP version range: positions
// after the repeated parameter must still line up with the arguments.
function connect($host, $port)
{
    $a = fsockopen($host, $port, $errno, $errstr, timeout());
    $b = pfsockopen($host, $port, $errno, $errstr, floatval(ini_get('default_socket_timeout')));
    $c = fsockopen($host, $port, <warning descr="Pass a variable here: this parameter is taken by reference.">code()</warning>, $errstr, 1.0);
    $d = pfsockopen($host, $port, $errno, <warning descr="Pass a variable here: this parameter is taken by reference.">strval($host)</warning>, timeout());
    $e = fsockopen($host, $port, timeout: timeout(), error_code: <warning descr="Pass a variable here: this parameter is taken by reference.">code()</warning>);
    return [$a, $b, $c, $d, $e];
}

function seal($data, $key, $ldap)
{
    openssl_seal($data, $sealed, $ekeys, array_values($key), strtolower('AES-256-CBC'), <warning descr="Pass a variable here: this parameter is taken by reference.">strval($data)</warning>);
    ldap_exop_passwd($ldap, strval($data), strval($data), strval($data), <warning descr="Pass a variable here: this parameter is taken by reference.">array_values($key)</warning>);
    ldap_exop_passwd($ldap, strval($data), strval($data), strval($data), $controls);
}
