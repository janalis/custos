<?php
namespace Check {
    function strlen($s) { return 3; }

    function probe(string $s) {
        $a = strlen($s) > 0;
        $b = \Util\mb_strlen($s) == 0;
        $c = $s === '';
        return [$a, $b, $c];
    }
}

namespace {
    function probe(string $s) {
        return $s !== '';
    }
}
