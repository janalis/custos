<?php
$parts=parse_url('https://example.test/path'); $parts['host']='example.test'; echo $parts['host'];

function guarded_access($raw, $uid) {
$p=parse_url($raw);
isset($p['host']); empty($p['host']); echo $p['host'] ?? 'fallback';
$p['host']['nested']='value';
unset($p['host']);
}
