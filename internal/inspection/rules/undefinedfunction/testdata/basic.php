<?php
namespace App;

use function Lib\Text\slug;

function local_helper() {}

<error descr="Fix the name or install the package declaring str_contians().">str_contians</error>('haystack', 'needle');
<error descr="Fix the name or install the package declaring strlenn().">strlenn</error>('x');
<error descr="Fix the name or install the package declaring App\Util\slugify().">Util\slugify</error>('Title');
<error descr="Fix the name or install the package declaring Lib\Text\slug().">slug</error>('Title');
<error descr="Fix the name or install the package declaring Other\run().">\Other\run</error>();

local_helper();
strlen('x');
\strlen('x');
array_map(null, []);

// A check whose result does not guard the call.
function unguarded()
{
    $loaded = extension_loaded('acme');
    <error descr="Fix the name or install the package declaring App\Acme\start().">Acme\start</error>($loaded);
}

// An earlier check about another value, name or extension.
function run($cb)
{
    if (!is_callable($cb) || !function_exists('other_start') || !method_exists($cb, 'run')) {
        return;
    }
    <error descr="Fix the name or install the package declaring App\Acme\stop().">Acme\stop</error>($cb);
}

// A conditional declaration about another function.
if (!function_exists('App\shimmed')) {
    function shimmed() { <error descr="Fix the name or install the package declaring helperOfShim().">helperOfShim</error>(); }
}
