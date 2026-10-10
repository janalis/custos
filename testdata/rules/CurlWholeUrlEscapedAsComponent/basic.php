<?php
<warning descr="Escape URL components before assembling the URL.">curl_setopt($ch, CURLOPT_URL, curl_escape($ch, 'https://example.net/a'))</warning>;
