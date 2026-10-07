<?php
function windows($long)
{
    $spec = <error descr="Malformed DateInterval specification.">'10D'</error>;
    $a = new DateInterval($spec);
    $b = new DateInterval($spec);
    $c = $long ? new DateInterval($spec) : null;

    $ok = 'P1D';
    $d = new DateInterval($ok);
    $e = new DateInterval($ok);

    $f = new DateInterval(<error descr="Malformed DateInterval specification.">'2W'</error>);
    $g = new DateInterval(<error descr="Malformed DateInterval specification.">'2W'</error>);
    return [$a, $b, $c, $d, $e, $f, $g];
}

function again()
{
    $spec = <error descr="Malformed DateInterval specification.">'10D'</error>;
    return [new DateInterval($spec), new DateInterval($spec)];
}
