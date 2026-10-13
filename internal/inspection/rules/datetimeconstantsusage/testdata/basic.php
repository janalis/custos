<?php
namespace Clock;

class Stamp { const ISO8601 = 'Y'; }
class LocalTime extends \DateTime {}
class Override extends \DateTime { const ISO8601 = 'x'; }

function formats(\DateTimeImmutable $at, LocalTime $lt)
{
    return [
        <error descr="DATE_ISO8601 is not ISO-8601 compliant; use DATE_ATOM.">DATE_ISO8601</error>,
        <error descr="DATE_ISO8601 is not ISO-8601 compliant; use DATE_ATOM.">\DATE_ISO8601</error>,
        <error descr="ISO8601 is not ISO-8601 compliant; use the ATOM constant.">\DateTime::ISO8601</error>,
        <error descr="ISO8601 is not ISO-8601 compliant; use the ATOM constant.">\DateTimeInterface::ISO8601</error>,
        <error descr="ISO8601 is not ISO-8601 compliant; use the ATOM constant.">LocalTime::ISO8601</error>,
        <error descr="ISO8601 is not ISO-8601 compliant; use the ATOM constant.">$lt::ISO8601</error>,
        Stamp::ISO8601,
        Override::ISO8601,
        DATE_ISO8601_LEGACY,
        date_iso8601,
        DATE_RFC3339,
        \DateTime::ATOM,
        $at->format('c'),
    ];
}
