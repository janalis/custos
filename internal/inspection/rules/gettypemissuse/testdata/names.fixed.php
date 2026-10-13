<?php
namespace Probe {
    function gettype($v) { return 'string'; }

    function check($v) {
        $a = gettype($v) === 'integer';
        $b = gettype($v) === 'int';
        $c = is_string($v);
        return [$a, $b, $c];
    }
}

namespace {
    function check($v) {
        return !is_array($v);
    }
}
