<?php
$h=curl_init();curl_setopt($h,CURLOPT_HEADER,true);curl_setopt($h,CURLOPT_RETURNTRANSFER,true);<warning descr="Keep HTTP headers separate from the decoded JSON body.">json_decode(curl_exec($h))</warning>;
