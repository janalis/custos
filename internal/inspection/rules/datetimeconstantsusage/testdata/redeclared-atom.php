<?php
namespace Clock;

class Year extends \DateTime { const ATOM = 'Y'; }
class Later extends Year {}
class Plain extends \DateTime {}

function formats(Year $y, Plain $p)
{
    return [
        // ATOM is redeclared on the receiver: reported, no fix.
        <error descr="ISO8601 is not ISO-8601 compliant; use the ATOM constant.">Year::ISO8601</error>,
        <error descr="ISO8601 is not ISO-8601 compliant; use the ATOM constant.">Later::ISO8601</error>,
        <error descr="ISO8601 is not ISO-8601 compliant; use the ATOM constant.">$y::ISO8601</error>,
        // Built-in ATOM: fixed.
        <error descr="ISO8601 is not ISO-8601 compliant; use the ATOM constant.">Plain::ISO8601</error>,
        <error descr="ISO8601 is not ISO-8601 compliant; use the ATOM constant.">$p::ISO8601</error>,
    ];
}
