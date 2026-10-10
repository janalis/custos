<?php
$f = new NumberFormatter("en_US", NumberFormatter::DECIMAL); $n = $f->parse("37tail", NumberFormatter::TYPE_DOUBLE, $end); <warning descr="Reject unparsed number suffixes.">echo $n</warning>;
