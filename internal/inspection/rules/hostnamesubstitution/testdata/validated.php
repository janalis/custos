<?php
// A list test before the assignment, on another variable or in a nested
// closure does not validate the stored host.
function before(array $allowed, $host)
{
    $ok = in_array($host, $allowed, true);
    $host = $_SERVER['HTTP_HOST'];
    return 'noreply@' . <error descr="E-mail address built from client-controlled $_SERVER['HTTP_HOST']; validate it against a whitelist.">$host</error>;
}

function other(array $allowed, $name)
{
    $host = $_SERVER['SERVER_NAME'];
    if (in_array($name, $allowed, true) || in_array(...) || in_array($$name, $allowed)) {
        return 'noreply@' . <error descr="E-mail address built from client-controlled $_SERVER['SERVER_NAME']; validate it against a whitelist.">$host</error>;
    }
    return null;
}

function nested(array $allowed)
{
    $host = $_SERVER['HTTP_HOST'];
    $check = function ($host) use ($allowed) {
        return in_array($host, $allowed, true);
    };
    return 'noreply@' . <error descr="E-mail address built from client-controlled $_SERVER['HTTP_HOST']; validate it against a whitelist.">$host</error>;
}

// String-returning and unknown calls still carry the host name.
$mailHost = strtolower(<error descr="Client-controlled host name stored here; validate it against a whitelist.">$_SERVER['HTTP_HOST']</error>);
$siteDomain = -normalise(<error descr="Client-controlled host name stored here; validate it against a whitelist.">$_SERVER['SERVER_NAME']</error>) + 0;
$from = 'user@' . <error descr="E-mail address built from client-controlled $_SERVER['HTTP_HOST']; validate it against a whitelist.">strlen($_SERVER['HTTP_HOST'])</error>;
