<?php
use IntlGregorianCalendar as BuiltinClass2;
$c = new BuiltinClass2(); <warning descr="Use zero-based Gregorian month numbers.">$c->set(2026, 12, 18)</warning>;
