<?php
namespace Calendar;

use DateInterval as Span;

class DateInterval { public function __construct($s) {} }

function f()
{
    return [new DateInterval('bogus'), new Span(<error descr="Malformed DateInterval specification.">'bogus'</error>)];
}
