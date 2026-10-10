<?php
$f = static function () {}; $bound = $f->bindTo(new stdClass()); <error descr="Check closure binding before invoking it.">$bound()</error>;
