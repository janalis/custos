<?php
$f = new NumberFormatter("en_US", NumberFormatter::CURRENCY); <error descr="Parse currency with parseCurrency.">$f->parse("$37.20", NumberFormatter::TYPE_CURRENCY)</error>;
