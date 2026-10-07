<?php
namespace Check {
    function strlen($s) { return 3; }

    function probe(string $s) {
        $a = strlen($s) > 0;
        $b = \Util\mb_strlen($s) == 0;
        $c = <weak_warning descr="Compare with an empty string instead: '$s === '''.">\StrLen($s) === 0</weak_warning>;
        return [$a, $b, $c];
    }
}

namespace {
    function probe(string $s) {
        return <weak_warning descr="Compare with an empty string instead: '$s !== '''.">MB_STRLEN($s) > 0</weak_warning>;
    }
}
