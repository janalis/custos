<?php
$routes = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">require_once __DIR__ . '/routes.php'</error>;
$ok = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $first ||
    include_once $second</error>;
register((<error descr="Only the first include_once/require_once returns the file's value; later ones return true.">require_once $plugin</error>));
$list = [<error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once 'menu.php'</error>];
$same = (<error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once 'a.php'</error>) === true;
$loose = (<error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once 'a.php'</error>) == 1;
require_once __DIR__ . '/bootstrap.php';
include_once('helpers.php');
$cfg = (require 'config.php');
