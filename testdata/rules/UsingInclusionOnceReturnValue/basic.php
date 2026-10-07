<?php
$routes = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">require_once __DIR__ . '/routes.php'</error>;
$ok = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $first ||
    <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $second</error></error>;
while (<error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once($plugin)</error>) {
    register((<error descr="Only the first include_once/require_once returns the file's value; later ones return true.">require_once $plugin</error>));
}
require_once __DIR__ . '/bootstrap.php';
include_once('helpers.php');
$cfg = (require 'config.php');
