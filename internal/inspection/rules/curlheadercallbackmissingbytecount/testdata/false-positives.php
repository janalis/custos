<?php
curl_setopt($ch, CURLOPT_HEADERFUNCTION, function($ch, $line) { return strlen($line); });
