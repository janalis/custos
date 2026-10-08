<?php
namespace Paths {
    function dirname($path) { return $path; }

    $a = \dirname(\dirname(__DIR__)) . '/conf';
}

namespace Imported {
    use function Paths\dirname;

    $b = \dirname(__DIR__);
}

namespace Plain {
    $c = dirname(__DIR__);
}
