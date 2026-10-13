<?php
function legacyIntervals()
{
    $mixed = <error descr="Weeks and days cannot be combined in a DateInterval specification before PHP 8.0.">'P3W1D'</error>;

    return [
        new DateInterval(<error descr="Weeks and days cannot be combined in a DateInterval specification before PHP 8.0.">'P1W2D'</error>),
        new DateInterval(<error descr="Weeks and days cannot be combined in a DateInterval specification before PHP 8.0.">'P1Y2M3W4DT5H'</error>),
        new DateInterval(<error descr="Malformed DateInterval specification.">'P2D1W'</error>),
        new DateInterval($mixed),
        new DateInterval($mixed),
        new DateInterval('P2W'),
        new DateInterval('P2WT3H'),
        new DateInterval('P5D'),
        new DateInterval('P1DT2H'),
    ];
}
