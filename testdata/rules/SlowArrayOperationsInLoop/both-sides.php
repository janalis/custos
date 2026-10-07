<?php
for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop."><error descr="'strlen(...)' is re-evaluated on every iteration; compute it once before the loop.">count($a) < strlen($b)</error></error>; $i++) {}
