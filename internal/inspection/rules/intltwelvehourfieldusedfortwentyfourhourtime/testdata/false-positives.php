<?php
function decremented(IntlCalendar $c) { $h = 12; $h--; $c->set(IntlCalendar::FIELD_HOUR, $h); }
function incremented(IntlCalendar $c) { $h = 23; ++$h; $c->set(IntlCalendar::FIELD_HOUR, $h); }
function captured(IntlCalendar $c) { $h = 12; $g = function () use (&$h) { $h = 3; }; $g(); $c->set(IntlCalendar::FIELD_HOUR, $h); }
function caught(IntlCalendar $c) { $h = 12; try { risky(); } catch (Exception $h) {} $c->set(IntlCalendar::FIELD_HOUR, $h); }
function compound(IntlCalendar $c) { $h = 12; $h -= 3; $c->set(IntlCalendar::FIELD_HOUR, $h); }
function twelveHour(IntlCalendar $c) { $c->set(IntlCalendar::FIELD_HOUR, 11); }
