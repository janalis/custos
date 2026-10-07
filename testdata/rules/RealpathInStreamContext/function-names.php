<?php
namespace Deploy\Paths {
    // user helper that understands phar:// paths itself
    function realpath($path) { return $path; }

    class Locator
    {
        public function home() { return realpath(__DIR__ . '/../home'); }
        public function etc() { return Fs\realpath(__DIR__ . '/../etc'); }
        public function bin() { return <warning descr="Use 'dirname(__DIR__) . '/bin'' instead: realpath() fails inside stream wrappers.">\RealPath(__DIR__ . '/../bin')</warning>; }
    }
}

namespace {
    use function Deploy\Paths\realpath;

    $a = realpath(__DIR__ . '/../a');
    $b = <warning descr="Use 'dirname(__DIR__) . '/b'' instead: realpath() fails inside stream wrappers.">\REALPATH(__DIR__ . '/../b')</warning>;
}
