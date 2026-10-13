<?php
foreach ($rows as $i => $row) {
    log_row($row);
    $a[tag($i++, $row)] = tag($i++, $row);
    $b[wrap(fn () => 1)] = <warning descr="This call is repeated in the key; store its result in a local variable.">wrap(fn () => 1)</warning>;
    $c[$fn($row)] = <warning descr="This call is repeated in the key; store its result in a local variable.">$fn($row)</warning>;
}
