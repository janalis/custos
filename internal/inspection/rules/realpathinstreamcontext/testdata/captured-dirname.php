<?php
namespace Paths {
    function dirname($path) { return $path; }

    $a = <warning descr="Use '\dirname(\dirname(__DIR__)) . '/conf'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../../conf')</warning>;
}

namespace Imported {
    use function Paths\dirname;

    $b = <warning descr="Use '\dirname(__DIR__)' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../')</warning>;
}

namespace Plain {
    $c = <warning descr="Use 'dirname(__DIR__)' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../')</warning>;
}
