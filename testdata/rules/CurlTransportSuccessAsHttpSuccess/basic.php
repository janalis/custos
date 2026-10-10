<?php
function bad($url) { $ch = curl_init('https://example.invalid/api'); if (<warning descr="Check the HTTP response status before reporting success.">curl_exec($ch)</warning> !== false) { return true; } return false; }
