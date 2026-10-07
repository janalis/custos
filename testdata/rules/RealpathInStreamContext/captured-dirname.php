<?php
namespace Paths {
    function dirname($path) { return $path; }

    $a = <warning descr="Use '\dirname(\dirname(__DIR__)) . '/conf'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../../conf')</warning>;
}

namespace Imported {
    use function Paths\dirname;

    $b = <warning descr="Use '\dirname($root) . '/'' instead: realpath() fails inside stream wrappers.">realpath($root . '/../')</warning>;
}

namespace Plain {
    $c = <warning descr="Use 'dirname($root) . '/'' instead: realpath() fails inside stream wrappers.">realpath($root . '/../')</warning>;
}
