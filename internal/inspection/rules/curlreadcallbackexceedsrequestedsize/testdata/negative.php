<?php
curl_setopt($ch, CURLOPT_READFUNCTION, fn($ch, $h, $limit) => str_repeat('x', $limit));
