<?php
namespace Scheduler {
    use DATEINTERVAL as Every;

    function plan()
    {
        return [
            new \dateinterval(<error descr="Malformed DateInterval specification.">'3M'</error>),
            new Every(<error descr="Malformed DateInterval specification.">'PT'</error>),
            new \DateINTERVAL('P1M'),
            new dateinterval('bogus'),
        ];
    }
}
