<?php
<error descr="Match the callback parameters to the supplied arrays.">array_map(fn($value, $key) => $key . $value, $items)</error>;

function join_pair($left, $right) { return $left . $right; }
<error descr="Match the callback parameters to the supplied arrays.">array_map('join_pair', $items)</error>;
