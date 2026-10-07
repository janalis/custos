<?php
namespace Audit\Tokens {
    // namespace-local helper shadowing the builtin: arguments may repeat
    function strcmp($a, $b) { return 0; }

    function check(string $t) {
        $x = strcmp($t, $t);
        $y = <error descr="Both compared strings are the same expression; one of them is probably wrong.">\StrCmp($t, $t)</error>;
        $z = Other\hash_equals($t, $t);
        return [$x, $y, $z];
    }
}

namespace Audit\Checks {
    use function Audit\Tokens\strcmp;

    function again(string $t) {
        $u = strcmp($t, $t);
        $v = <error descr="Both compared strings are the same expression; one of them is probably wrong.">SubStr_Compare($t, $t, 0)</error>;
        return [$u, $v];
    }
}
