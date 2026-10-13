<?php
use IntlCalendar as BuiltinClass1;
$c = BuiltinClass1::createInstance(); <warning descr="Convert seconds to calendar milliseconds.">$c->setTime(time())</warning>;
