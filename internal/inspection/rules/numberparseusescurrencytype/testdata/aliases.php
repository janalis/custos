<?php
use NumberFormatter as BuiltinClass0;
$f = new BuiltinClass0("en_US", BuiltinClass0::CURRENCY); <error descr="Parse currency with parseCurrency.">$f->parse("$37.20", BuiltinClass0::TYPE_CURRENCY)</error>;
