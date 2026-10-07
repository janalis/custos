<?php
namespace Clock;

class Stamp { const ISO8601 = 'Y'; }
class LocalTime extends \DateTime {}
class Override extends \DateTime { const ISO8601 = 'x'; }

function formats(\DateTimeImmutable $at, LocalTime $lt)
{
    return [
        DATE_ATOM,
        \DATE_ATOM,
        \DateTime::ATOM,
        \DateTimeInterface::ATOM,
        LocalTime::ATOM,
        $lt::ATOM,
        Stamp::ISO8601,
        Override::ISO8601,
        DATE_ISO8601_LEGACY,
        date_iso8601,
        DATE_RFC3339,
        \DateTime::ATOM,
        $at->format('c'),
    ];
}
