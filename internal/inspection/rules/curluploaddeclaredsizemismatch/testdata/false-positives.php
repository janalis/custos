<?php
$h = fopen('data://text/plain,abcdef', 'r'); curl_setopt($ch, CURLOPT_UPLOAD, true); curl_setopt($ch, CURLOPT_INFILE, $h); curl_setopt($ch, CURLOPT_INFILESIZE, 6);
