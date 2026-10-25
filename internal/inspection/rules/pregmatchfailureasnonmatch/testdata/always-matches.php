<?php
function utf8(string $s, string $p): bool
{
    if (!preg_match('//u', $s)) { return false; }
    if (!preg_match('{^}u', $s)) { return false; }
    if (!<warning descr="Distinguish regex failure from a successful non-match.">preg_match('/./su', $s)</warning>) { return false; }
    if (!<warning descr="Distinguish regex failure from a successful non-match.">preg_match('/', $s)</warning>) { return false; }
    if (!<warning descr="Distinguish regex failure from a successful non-match.">preg_match('a', $s)</warning>) { return false; }
    if (!<warning descr="Distinguish regex failure from a successful non-match.">preg_match('{x', $s)</warning>) { return false; }
    return true;
}
