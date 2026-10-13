<?php
$sender  = 'noreply@' . <error descr="E-mail address built from client-controlled $_SERVER['HTTP_HOST']; validate it against a whitelist.">$_SERVER['HTTP_HOST']</error>;
$bounce  = 'x' . 'bounce@' . <error descr="E-mail address built from client-controlled $_SERVER['SERVER_NAME']; validate it against a whitelist.">trim($_SERVER["SERVER_NAME"])</error> . '.local';
$siteHost = <error descr="Client-controlled host name stored here; validate it against a whitelist.">$_SERVER['SERVER_NAME']</error>;

function contact()
{
    $base = strtolower($_SERVER['HTTP_HOST']);
    $base = str_replace('www.', '', $base);
    return 'help@' . <error descr="E-mail address built from client-controlled $_SERVER['HTTP_HOST']; validate it against a whitelist.">$base</error>;
}

class Mailer
{
    public function configure()
    {
        $this->mailDomain = <error descr="Client-controlled host name stored here; validate it against a whitelist.">$_SERVER['HTTP_HOST']</error>;
        static::$EmailFrom = <error descr="Client-controlled host name stored here; validate it against a whitelist.">$_SERVER['SERVER_NAME']</error>;
    }
}

function derived()
{
    $base = $_SERVER['HTTP_HOST'];
    $base = strtolower($base);
    return 'help@' . <error descr="E-mail address built from client-controlled $_SERVER['HTTP_HOST']; validate it against a whitelist.">$base</error>;
}

function conditionalOverwrite($flag)
{
    $base = $_SERVER['HTTP_HOST'];
    if ($flag) {
        $base = 'example.org';
    }
    return 'help@' . <error descr="E-mail address built from client-controlled $_SERVER['HTTP_HOST']; validate it against a whitelist.">$base</error>;
}

function usedBeforeOverwrite()
{
    $base = $_SERVER['HTTP_HOST'];
    $first = 'help@' . <error descr="E-mail address built from client-controlled $_SERVER['HTTP_HOST']; validate it against a whitelist.">$base</error>;
    $base = 'example.org';
    return $first . 'help@' . $base;
}
