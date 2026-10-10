<?php
$c=IntlCalendar::createInstance();$t=time();<warning descr="Convert seconds to calendar milliseconds.">$c->setTime($t)</warning>;
