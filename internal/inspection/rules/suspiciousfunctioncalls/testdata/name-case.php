<?php
function sameSide(string $secret, string $given) {
    $p = <error descr="Both compared strings are the same expression; one of them is probably wrong.">STRCMP($given, $given)</error>;
    $q = <error descr="Both compared strings are the same expression; one of them is probably wrong.">\Hash_Equals($secret, $secret)</error>;
    $r = StrNatCaseCmp($secret, $given);
    return [$p, $q, $r];
}
