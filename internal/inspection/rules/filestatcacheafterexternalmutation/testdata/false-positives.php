<?php
// From PHP 8.3 on, exec() and system() clear the stat cache themselves.
$before = filesize("/tmp/cache.bin");
exec("truncate -s 0 /tmp/cache.bin");
$after = filesize("/tmp/cache.bin");

$before = filemtime("/tmp/cache.bin");
system("touch /tmp/cache.bin");
$after = filemtime("/tmp/cache.bin");
