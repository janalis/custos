<?php
// The requested size may be changed before the return: no proof.
function clamp(&$n) { $n = max(0, $n - 16); }
$ch = curl_init();
curl_setopt($ch, CURLOPT_READFUNCTION, function ($c, $h, $limit) { clamp($limit); return str_repeat('x', $limit + 1); });
curl_setopt($ch, CURLOPT_READFUNCTION, function ($c, $h, $limit) { foreach ([$limit - 8] as $limit) {} return str_repeat('x', $limit + 1); });
curl_setopt($ch, CURLOPT_READFUNCTION, function ($c, $h, $limit) { settype($limit, 'string'); return str_repeat('x', $limit + 1); });
curl_setopt($ch, CURLOPT_READFUNCTION, function ($c, $h, $limit) { extract(['limit' => 0]); return str_repeat('x', $limit + 1); });
curl_setopt($ch, CURLOPT_READFUNCTION, function ($c, $h, $limit) { [$limit] = [0]; return str_repeat('x', $limit + 1); });
curl_setopt($ch, CURLOPT_READFUNCTION, function ($c, $h, $limit) { $alias = &$limit; $alias = 0; return str_repeat('x', $limit + 1); });
curl_setopt($ch, CURLOPT_READFUNCTION, function ($c, $h, $limit) { $shrink = function () use (&$limit) { $limit = 0; }; $shrink(); return str_repeat('x', $limit + 1); });
curl_setopt($ch, CURLOPT_READFUNCTION, function ($c, $h, $limit) { $name = 'limit'; $$name = 0; return str_repeat('x', $limit + 1); });
