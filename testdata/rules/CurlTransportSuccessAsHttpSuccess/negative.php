<?php
function good($url) { $ch = curl_init('https://example.invalid/api'); curl_setopt($ch, CURLOPT_FAILONERROR, true); return curl_exec($ch) !== false; }
