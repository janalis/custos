<?php
namespace Paths {
    function dirname($path) { return $path; }

    $a = \dirname(\dirname(__DIR__)) . '/conf';
}

namespace Imported {
    use function Paths\dirname;

    $b = \dirname($root);
}

namespace Plain {
    $c = dirname($root);
}
