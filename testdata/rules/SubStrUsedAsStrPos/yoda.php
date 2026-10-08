<?php
$r = <weak_warning descr="Use '0 === strpos($uri, $base)' instead.">substr($uri, 0, strlen($base)) === $base</weak_warning>;
