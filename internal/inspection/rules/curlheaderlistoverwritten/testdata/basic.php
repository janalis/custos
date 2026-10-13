<?php
$h=curl_init();curl_setopt($h,CURLOPT_HTTPHEADER,["Authorization: Bearer x"]);curl_setopt($h,CURLOPT_HTTPHEADER,<warning descr="Combine HTTP headers before setting the final list.">["Accept: application/json"]</warning>);curl_exec($h);
