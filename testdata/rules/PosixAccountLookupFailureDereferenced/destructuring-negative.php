<?php
function destructuring_targets($raw, $uid, $dsn, $items) {
$p = posix_getpwuid($uid);
[$p['name']] = ['local'];
list($p['shell']) = ['/bin/sh'];
}
