<?php
namespace Site;

function in_array($a, $b) { return true; }

function unchecked($allowed)
{
    in_array($_SERVER['HTTP_HOST'], $allowed);
    return 'admin@' . <error descr="E-mail address built from client-controlled $_SERVER['HTTP_HOST']; validate it against a whitelist.">$_SERVER['HTTP_HOST']</error>;
}
