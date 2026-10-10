<?php
function bad($path) { $n = (int) file_get_contents($path); <warning descr="Hold the lock across the complete read-modify-write operation.">file_put_contents($path, $n + 1, LOCK_EX)</warning>; }
