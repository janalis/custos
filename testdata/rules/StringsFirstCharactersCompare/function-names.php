<?php
namespace Text\Compare {
    // user function: third argument is not a byte length
    function strncmp($a, $b, $words) { return 0; }

    function starts($s) {
        return [
            strncmp($s, 'abc', 2),
            Legacy\strncasecmp($s, 'abc', 2),
            \StrNCmp($s, 'abc', <error descr="Length 2 does not match the 3-character literal.">2</error>),
        ];
    }
}

namespace {
    use function Text\Compare\strncmp;

    $a = strncmp($s, 'abcd', 2);
    $b = STRNCASECMP($s, 'abcd', <error descr="Length 2 does not match the 4-character literal.">2</error>);
}
