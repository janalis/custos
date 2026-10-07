<?php
namespace {
    <warning descr="Use opcache_compile_file() in a preload script instead of require.">require __DIR__ . '/src/Kernel.php'</warning>;
    <warning descr="Use opcache_compile_file() in a preload script instead of include_once.">include_once($base . 'helpers.php')</warning>;
    require_once ( 'vendor/autoload.php' );
    require "$root/vendor/autoload.php";
    require dirname(__DIR__) . '/var/cache/prod/App_KernelProdContainer.preload.php';
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

namespace Broken {
    require;
}
