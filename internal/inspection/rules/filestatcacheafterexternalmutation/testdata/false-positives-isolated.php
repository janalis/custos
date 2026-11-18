<?php
$before=filesize("/tmp/cache.bin");exec("touch /tmp/other.bin");$after=filesize("/tmp/cache.bin");$before=filesize("/tmp/cache.bin");exec("touch /tmp/cache.bin");clearstatcache();$after=filesize("/tmp/cache.bin");filesize($unknown);strlen("stat");
