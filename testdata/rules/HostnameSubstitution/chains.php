<?php
$grouped = 'admin@' . (<error descr="E-mail address built from client-controlled $_SERVER['HTTP_HOST']; validate it against a whitelist.">$_SERVER['HTTP_HOST']</error> . '.example');
$leftOnly = $_SERVER['HTTP_HOST'] . '@mail';

function grouped()
{
    $host = $_SERVER['SERVER_NAME'];
    $a = 'info@' . (<error descr="E-mail address built from client-controlled $_SERVER['SERVER_NAME']; validate it against a whitelist.">$host</error> . '.net');
    $b = 'sales@' . (<error descr="E-mail address built from client-controlled $_SERVER['SERVER_NAME']; validate it against a whitelist.">$host</error>);
    $c = $host . '@relay';
    $d = 'Host: ' . $host . ' ok';
    return [$a, $b, $c, $d];
}
