<?php
$h=curl_init();curl_setopt($h,CURLOPT_POSTFIELDS,["id"=>1]);curl_setopt($h,CURLOPT_HTTPHEADER,<warning descr="Encode the request body as JSON before declaring JSON content.">["Content-Type: application/json"]</warning>);curl_exec($h);
