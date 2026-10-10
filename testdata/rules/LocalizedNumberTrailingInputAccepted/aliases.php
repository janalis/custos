<?php
use NumberFormatter as BuiltinClass0;
$f = new BuiltinClass0("en_US", BuiltinClass0::DECIMAL); $n = $f->parse("37tail", BuiltinClass0::TYPE_DOUBLE, $end); <warning descr="Reject unparsed number suffixes.">echo $n</warning>;
