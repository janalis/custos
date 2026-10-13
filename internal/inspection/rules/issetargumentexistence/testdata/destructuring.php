<?php
function lazy(array $parts): string
{
    $out = '';
    foreach ($parts as $p) {
        if (!isset($ts)) {
            [$ts, $tz] = [time(), 'UTC'];
        }
        if (!isset($nested)) {
            list(, list($nested)) = [1, [2]];
        }
        if (!isset(<error descr="Variable '$other' is not defined in this scope.">$other</error>)) {
            [$a, ($b)] = [1, 2];
        }
        $out .= $ts . $tz . $nested . $p;
    }
    return $out;
}
