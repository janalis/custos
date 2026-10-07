<?php
function intervals($mode)
{
    $weekly = 'P1W';
    $broken = <error descr="Malformed DateInterval specification.">'1W'</error>;

    return [
        new DateInterval(<error descr="Malformed DateInterval specification.">'3M'</error>),
        new DateInterval(<error descr="Malformed DateInterval specification.">'P2D1Y'</error>),
        new DateInterval(<error descr="Malformed DateInterval specification.">'PT1D'</error>),
        new DateInterval(<error descr="Malformed DateInterval specification.">"PT"</error>),
        new DateInterval(<error descr="Malformed DateInterval specification.">'PT5'</error>),
        new DateInterval(<error descr="Malformed DateInterval specification.">'PT1.5S'</error>),
        new DateInterval(<error descr="Malformed DateInterval specification.">'P1W2D'</error>),
        new DateInterval(<error descr="Malformed DateInterval specification.">'PT10:30:00'</error>),
        new DateInterval(<error descr="Malformed DateInterval specification.">'p1d'</error>),
        new \DateInterval(<error descr="Malformed DateInterval specification.">'1D'</error>),
        new DateInterval($weekly),
        new DateInterval($broken),
        new DateInterval("P{$mode}D"),
        new DateInterval('P'),
        new DateInterval('P1Y2M3DT4H5M6S'),
        new DateInterval('PT36H'),
        new DateInterval('P1D', 'x'),
        new DateInterval('P0002-00-10T12:00:00'),
        new DateTimeImmutable('3M'),
        new DateInterval($mode ? 'P1D' : 'P2D'),
    ];
}

function later()
{
    $spec = <error descr="Malformed DateInterval specification.">'10D'</error>;
    return new DateInterval($spec);
}
