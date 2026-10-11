<?php
namespace Clock;

class Year extends \DateTime { const ATOM = 'Y'; }
class Later extends Year {}
class Plain extends \DateTime {}

function formats(Year $y, Plain $p)
{
    return [
        // ATOM is redeclared on the receiver: reported, no fix.
        Year::ISO8601,
        Later::ISO8601,
        $y::ISO8601,
        // Built-in ATOM: fixed.
        Plain::ATOM,
        $p::ATOM,
    ];
}
