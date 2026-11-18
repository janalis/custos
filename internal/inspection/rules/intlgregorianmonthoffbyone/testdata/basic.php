<?php
$c = new IntlGregorianCalendar(); <warning descr="Use zero-based Gregorian month numbers.">$c->set(2026, 12, 18)</warning>;
