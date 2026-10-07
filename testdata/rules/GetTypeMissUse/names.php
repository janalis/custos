<?php
namespace Probe {
    function gettype($v) { return 'string'; }

    function check($v) {
        $a = gettype($v) === 'integer';
        $b = gettype($v) === 'int';
        $c = <warning descr="Use 'is_string($v)' instead.">\GetType($v) === 'string'</warning>;
        return [$a, $b, $c];
    }
}

namespace {
    function check($v) {
        return <warning descr="Use '!is_array($v)' instead.">GETTYPE($v) !== 'array'</warning>;
    }
}
