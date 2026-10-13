<?php
$h=curl_init(); $cb=function($h,$stream,$length) {--$length; return str_repeat('x',$length+1);}; curl_setopt($h,CURLOPT_READFUNCTION,$cb); echo strlen($cb($h,null,10));
