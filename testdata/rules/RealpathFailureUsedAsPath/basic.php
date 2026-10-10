<?php
$root = realpath($path); <warning descr="Reject canonicalization failure before joining paths.">file_get_contents($root . '/config.json')</warning>;
