<?php
function unchanged(IntlCalendar $c) { $h = 19; $c->set(IntlCalendar::FIELD_HOUR_OF_DAY, $h); }
function copied(IntlCalendar $c) { $h = 19; $g = function () use ($h) { $h--; }; $c->set(IntlCalendar::FIELD_HOUR_OF_DAY, $h); }
