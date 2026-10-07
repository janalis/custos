<?php
namespace Deploy\Paths {
    // user helper that understands phar:// paths itself
    function realpath($path) { return $path; }

    class Locator
    {
        public function home() { return realpath(__DIR__ . '/../home'); }
        public function etc() { return Fs\realpath(__DIR__ . '/../etc'); }
        public function bin() { return dirname(__DIR__) . '/bin'; }
    }
}

namespace {
    use function Deploy\Paths\realpath;

    $a = realpath(__DIR__ . '/../a');
    $b = dirname(__DIR__) . '/b';
}
