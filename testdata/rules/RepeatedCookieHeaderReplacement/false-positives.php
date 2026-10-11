<?php

function sameCookieExpired()
{
    header('Set-Cookie: sid=abc; Path=/');
    header('Set-Cookie: sid=deleted; Max-Age=0; Path=/');
}

function sameCookieNewValue()
{
    header('Set-Cookie: theme=light');
    header('set-cookie:theme = dark');
}

function cookieHeadersRemoved()
{
    header('Set-Cookie: a=1');
    header_remove('Set-Cookie');
    header('Set-Cookie: b=2');
}

function cookieHeadersRemovedAnyCase()
{
    header('Set-Cookie: a=1');
    header_remove('set-cookie');
    header('Set-Cookie: b=2');
}

function everyHeaderRemoved()
{
    header('Set-Cookie: a=1');
    header_remove();
    header('Set-Cookie: b=2');
}

function unknownHeaderRemoved($field)
{
    header('Set-Cookie: a=1');
    header_remove($field);
    header('Set-Cookie: b=2');
}

function maybeRemoved($reset)
{
    header('Set-Cookie: a=1');
    if ($reset) {
        header_remove('Set-Cookie');
    }
    header('Set-Cookie: b=2');
}

function removedInsideLoop(array $pairs)
{
    foreach ($pairs as $pair) {
        header('Set-Cookie: a=1');
        header('Set-Cookie: b=2', false);
        header_remove('Set-Cookie');
    }
}

function namelessCookieLine()
{
    header('Set-Cookie: a=1');
    header('Set-Cookie: flag; Path=/a=b');
}

function appended()
{
    header('Set-Cookie: a=1');
    header('Set-Cookie: b=2', false);
}

function viaSetcookie()
{
    setcookie('a', '1');
    header('Set-Cookie: b=2');
}
