<?php
$c = IntlCalendar::createInstance(); <warning descr="Use the 24-hour calendar field.">$c->set(IntlCalendar::FIELD_HOUR, 19)</warning>;
