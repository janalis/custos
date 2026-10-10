<?php
$before=filesize("/tmp/cache.bin");exec("touch /tmp/cache.bin");$after=<warning descr="Clear cached file metadata after external file changes.">filesize("/tmp/cache.bin")</warning>;
