<?php
$ch=curl_init("http://api.example.test"); <warning descr="Protect basic authentication with HTTPS.">curl_setopt($ch,CURLOPT_USERPWD,"user:secret")</warning>; curl_exec($ch);
