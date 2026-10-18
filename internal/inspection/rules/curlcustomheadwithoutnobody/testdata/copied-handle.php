<?php
$handle = curl_init();
$copy = $handle;
curl_setopt($handle, CURLOPT_CUSTOMREQUEST, 'HEAD');
curl_exec($handle);
