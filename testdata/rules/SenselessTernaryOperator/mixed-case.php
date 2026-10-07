<?php
$m = <warning descr="This ternary always yields 'Count($b)'; use it directly.">count($a) === COUNT($b) ? Count($a) : Count($b)</warning>;
$n = $box->Size === $limit ? $box->size : $limit;
