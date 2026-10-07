<?php
namespace {
    <warning descr="Use opcache_compile_file() in a preload script instead of require.">require __DIR__ . '/src/Kernel.php'</warning>;
    <warning descr="Use opcache_compile_file() in a preload script instead of include_once.">include_once($base . 'helpers.php')</warning>;
    <warning descr="Use opcache_compile_file() in a preload script instead of require_once.">require_once ( 'vendor/autoload.php' )</warning>;
    foreach ($files as $file) {
        <warning descr="Use opcache_compile_file() in a preload script instead of include.">include $file</warning>;
    }
    $map = require 'classmap.php';
    $cfg = (require 'config.php');
    return include 'tail.php';
    if (include 'x.php') {}
}

namespace Boot {
    use function Opcache\opcache_compile_file;

    <warning descr="Use opcache_compile_file() in a preload script instead of require.">require 'boot.php'</warning>;
}

namespace Warm {
    function opcache_compile_file($file) { return true; }

    <warning descr="Use opcache_compile_file() in a preload script instead of include.">include 'warm.php'</warning>;
}
