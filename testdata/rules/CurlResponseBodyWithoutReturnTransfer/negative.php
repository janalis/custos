<?php
$ch = curl_init('https://example.invalid/'); curl_setopt($ch, CURLOPT_RETURNTRANSFER, true); $body = curl_exec($ch); json_decode($body, true);
