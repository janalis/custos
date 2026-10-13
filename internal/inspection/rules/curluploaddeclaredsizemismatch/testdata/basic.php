<?php
$h = fopen('data://text/plain,abcdef', 'r'); curl_setopt($ch, CURLOPT_UPLOAD, true); curl_setopt($ch, CURLOPT_INFILE, $h); <warning descr="Match the declared upload size to remaining bytes.">curl_setopt($ch, CURLOPT_INFILESIZE, 99)</warning>;
