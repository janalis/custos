<?php
function destructuring_targets($raw, $uid, $dsn, $items) {
$p = parse_url($raw);
[$p['host']] = ['example.test'];
list($p['port']) = [443];
foreach ($items as [$p['path']]) {}
}
