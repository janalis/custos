<?php
function parenthesized(IntlCalendar $c) {
    <warning descr="Convert seconds to calendar milliseconds.">$c->setTime((/* wall clock, see OPS-12 */ time()))</warning>;
}
