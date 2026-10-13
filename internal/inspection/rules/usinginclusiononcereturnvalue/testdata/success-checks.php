<?php
// include_once returns false only on failure: success tests are reliable.
if (!include_once __DIR__ . '/lib.php') {
    exit(1);
}
while (include_once($plugin)) {
    break;
}
$ready = (@include_once $file) !== false;
$failed = false === @include_once $file;
include_once 'a.php' or die('missing');
if ((include_once 'a.php') xor $other) {
}
$label = (include_once 'b.php') ? 'yes' : 'no';
if (include_once 'c.php' && $x) {
}
$found = (bool) include_once $path;
$also = (boolean) (@include_once $path);
$num = (int) <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $path</error>;
