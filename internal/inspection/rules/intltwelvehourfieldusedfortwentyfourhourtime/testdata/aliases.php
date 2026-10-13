<?php
use IntlCalendar as BuiltinClass1;
$c = BuiltinClass1::createInstance(); <warning descr="Use the 24-hour calendar field.">$c->set(BuiltinClass1::FIELD_HOUR, 19)</warning>;
