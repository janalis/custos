<?php
namespace App;

use function array_key_last as last;

if (<error descr="str_contains() needs PHP 8.0; the project targets PHP 5.6.">str_contains</error>($haystack, 'needle')) {
    echo <error descr="array_key_last() needs PHP 7.3; the project targets PHP 5.6.">\array_key_last</error>($items);
}
<error descr="array_key_last() needs PHP 7.3; the project targets PHP 5.6.">last</error>($items);
$n = <error descr="intdiv() needs PHP 7.0; the project targets PHP 5.6.">intdiv</error>(7, 2);
$bytes = <error descr="random_bytes() needs PHP 7.0; the project targets PHP 5.6.">random_bytes</error>(16);

// A check about something else guards nothing here.
if (function_exists('array_key_first')) {
    <error descr="is_countable() needs PHP 7.3; the project targets PHP 5.6.">is_countable</error>($items);
}
