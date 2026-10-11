<?php
function parenthesized(IntlCalendar $c) {
    $c->setTime((/* wall clock, see OPS-12 */ time() * 1000));
}
