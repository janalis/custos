<?php
$config = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">REQUIRE_ONCE __DIR__ . '/config.php'</error>;
$menu = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">Include_Once 'menu.php'</error>;
INCLUDE_ONCE 'side-effects.php';
$plain = REQUIRE 'plain.php';
