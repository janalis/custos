<?php
function viaGlobal(&$config)
{
    global $config;
    unset($config);
}

function viaStatic($cache, $other)
{
    static $cache;
    unset($cache);
    <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">unset($other);</weak_warning>
}

function beforeRebinding($state)
{
    <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">unset($state);</weak_warning>
    global $state;
    unset($state);
}

function nestedRebinding($p)
{
    $f = function () { global $p; };
    <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">unset($p);</weak_warning>
}
