<?php
$ch=curl_init("https://api.example.test"); curl_setopt($ch,CURLOPT_USERPWD,$credentials); curl_exec($ch);
