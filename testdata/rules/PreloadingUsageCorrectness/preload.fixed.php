<?php
namespace {
    opcache_compile_file(__DIR__ . '/src/Kernel.php');
    opcache_compile_file($base . 'helpers.php');
    require_once ( 'vendor/autoload.php' );
    require "$root/vendor/autoload.php";
    require dirname(__DIR__) . '/var/cache/prod/App_KernelProdContainer.preload.php';
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

namespace Broken {
    require;
}

namespace Espo {
    // A bootstrap whose effects the following code needs must run.
    include "bootstrap.php";
    use Espo\Core\Application;
    (new Application())->run();
}

namespace Calls {
    require 'a.php';
    helper();
    require 'b.php';
    Registry::warm();
    require 'c.php';
    $warmer->warm();
    require 'd.php';
    $fn();
    opcache_compile_file('e.php');
    strlen('x');
    function helper() { return new \ArrayObject(); }
    class Registry { public static function warm() { helper(); } }
    if ($ready) opcache_compile_file('f.php');
}
