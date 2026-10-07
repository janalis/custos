<?php
namespace {
    opcache_compile_file(__DIR__ . '/src/Kernel.php');
    opcache_compile_file($base . 'helpers.php');
    opcache_compile_file('vendor/autoload.php');
    foreach ($files as $file) {
        opcache_compile_file($file);
    }
    $map = require 'classmap.php';
    $cfg = (require 'config.php');
    return include 'tail.php';
    if (include 'x.php') {}
}

namespace Boot {
    use function Opcache\opcache_compile_file;

    \opcache_compile_file('boot.php');
}

namespace Warm {
    function opcache_compile_file($file) { return true; }

    \opcache_compile_file('warm.php');
}
