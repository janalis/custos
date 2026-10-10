<?php
$ch = curl_init('https://example.invalid/'); $body = curl_exec($ch); <warning descr="Enable response-body return before decoding curl output.">json_decode($body, true)</warning>;
