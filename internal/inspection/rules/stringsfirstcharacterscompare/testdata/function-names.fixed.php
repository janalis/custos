<?php
namespace Text\Compare {
    // user function: third argument is not a byte length
    function strncmp($a, $b, $words) { return 0; }

    function starts($s) {
        return [
            strncmp($s, 'abc', 2),
            Legacy\strncasecmp($s, 'abc', 2),
            \StrNCmp($s, 'abc', 3),
        ];
    }
}

namespace {
    use function Text\Compare\strncmp;

    $a = strncmp($s, 'abcd', 2);
    $b = STRNCASECMP($s, 'abcd', 4);
}
