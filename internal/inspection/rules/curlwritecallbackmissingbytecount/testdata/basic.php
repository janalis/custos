<?php
$h=curl_init();curl_setopt($h,CURLOPT_WRITEFUNCTION,<warning descr="Return the number of bytes handled by the write callback.">function($h,$data){echo $data;}</warning>);
curl_setopt($h,CURLOPT_WRITEFUNCTION,<warning descr="Return the number of bytes handled by the write callback.">function($h,$data){return null;}</warning>);
