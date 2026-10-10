<?php
$parts=posix_getpwuid(posix_getuid());$parts['name']='fallback';echo $parts['name'];

function guarded_access($raw, $uid) {
$p=posix_getpwuid($uid);
isset($p['name']); empty($p['name']); echo $p['name'] ?? 'fallback';
$p['name']['nested']='value';
unset($p['name']);
}
