<?php
if(<warning descr="Accept only verification result 1.">openssl_verify($s,$sig,$key,OPENSSL_ALGO_SHA256)</warning>){echo "accepted";}
