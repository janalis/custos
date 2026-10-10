<?php
curl_setopt($ch, CURLOPT_URL, 'https://example.net/' . curl_escape($ch, 'a b'));
