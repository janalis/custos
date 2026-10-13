<?php
$h=curl_init();curl_setopt($h,CURLOPT_FOLLOWLOCATION,true);curl_setopt($h,<warning descr="Restrict credential forwarding across redirects.">CURLOPT_UNRESTRICTED_AUTH</warning>,true);curl_setopt($h,CURLOPT_USERPWD,"user:secret");curl_exec($h);
