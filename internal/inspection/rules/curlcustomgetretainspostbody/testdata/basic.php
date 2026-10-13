<?php
$h=curl_init();curl_setopt($h,CURLOPT_POSTFIELDS,"x=1");curl_setopt($h,CURLOPT_CUSTOMREQUEST,"GET");<warning descr="Clear the request body before switching to GET.">curl_exec($h)</warning>;
$g=curl_init();curl_setopt($g,CURLOPT_POSTFIELDS,["x"=>1]);curl_setopt($g,CURLOPT_CUSTOMREQUEST,"GET");<warning descr="Clear the request body before switching to GET.">curl_exec($g)</warning>;
