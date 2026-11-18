<?php
$c = IntlCalendar::createInstance(); <warning descr="Convert seconds to calendar milliseconds.">$c->setTime(time())</warning>;
