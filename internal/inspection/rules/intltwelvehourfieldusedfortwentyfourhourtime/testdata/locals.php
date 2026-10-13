<?php
function unchanged(IntlCalendar $c) { $h = 19; <warning descr="Use the 24-hour calendar field.">$c->set(IntlCalendar::FIELD_HOUR, $h)</warning>; }
function copied(IntlCalendar $c) { $h = 19; $g = function () use ($h) { $h--; }; <warning descr="Use the 24-hour calendar field.">$c->set(IntlCalendar::FIELD_HOUR, $h)</warning>; }
